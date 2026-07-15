package com.mailx.data.remote.utils

import android.annotation.SuppressLint
import android.app.usage.NetworkStats
import android.app.usage.NetworkStatsManager
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.net.ConnectivityManager
import android.net.LinkProperties
import android.net.Network
import android.net.NetworkCapabilities
import android.net.NetworkInfo
import android.net.NetworkRequest
import android.net.TrafficStats
import android.net.wifi.ScanResult
import android.net.wifi.WifiConfiguration
import android.net.wifi.WifiInfo
import android.net.wifi.WifiManager
import android.os.BatteryManager
import android.os.Build
import android.os.Handler
import android.os.Looper
import android.os.PowerManager
import android.provider.Settings
import android.telephony.PhoneStateListener
import android.telephony.ServiceState
import android.telephony.SignalStrength
import android.telephony.TelephonyDisplayInfo
import android.telephony.TelephonyManager
import androidx.annotation.RequiresPermission
import androidx.core.content.getSystemService
import com.mailx.core.Logger
import com.mailx.domain.model.NetworkType
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.*
import java.io.BufferedReader
import java.io.InputStreamReader
import java.io.OutputStreamWriter
import java.net.HttpURLConnection
import java.net.InetAddress
import java.net.InetSocketAddress
import java.net.InterfaceAddress
import java.net.NetworkInterface
import java.net.Socket
import java.net.URL
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.ConcurrentLinkedDeque
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicBoolean
import java.util.concurrent.atomic.AtomicInteger
import java.util.concurrent.atomic.AtomicLong
import java.util.concurrent.atomic.AtomicReference
import javax.inject.Inject
import javax.inject.Singleton
import javax.net.SocketFactory
import javax.net.ssl.HttpsURLConnection
import javax.net.ssl.SSLContext
import javax.net.ssl.SSLHandshakeException
import javax.net.ssl.SSLSocket
import javax.net.ssl.TrustManager
import javax.net.ssl.X509TrustManager
import kotlin.math.abs
import kotlin.math.ln
import kotlin.math.pow
import kotlin.math.sqrt

@Singleton
class NetworkStateMonitor @Inject constructor(
    private val context: Context,
    private val logger: Logger
) {

    private val connectivityManager: ConnectivityManager?
        get() = context.getSystemService()

    private val wifiManager: WifiManager?
        get() = context.applicationContext.getSystemService()

    private val telephonyManager: TelephonyManager?
        get() = context.getSystemService()

    private val powerManager: PowerManager?
        get() = context.getSystemService()

    private val networkStatsManager: NetworkStatsManager?
        get() = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            context.getSystemService()
        } else null

    private val mainHandler = Handler(Looper.getMainLooper())
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    private val _networkState = MutableStateFlow(ComprehensiveNetworkState())
    val networkState: StateFlow<ComprehensiveNetworkState> = _networkState.asStateFlow()

    private val _isOnline = MutableStateFlow(false)
    val isOnline: StateFlow<Boolean> = _isOnline.asStateFlow()

    private val _networkType = MutableStateFlow(NetworkType.UNKNOWN)
    val networkType: StateFlow<NetworkType> = _networkType.asStateFlow()

    private val _bandwidthEstimate = MutableStateFlow(BandwidthEstimate.UNKNOWN)
    val bandwidthEstimate: StateFlow<BandwidthEstimate> = _bandwidthEstimate.asStateFlow()

    private val _connectionQuality = MutableStateFlow(ConnectionQuality.UNKNOWN)
    val connectionQuality: StateFlow<ConnectionQuality> = _connectionQuality.asStateFlow()

    private val _diagnosticResult = MutableStateFlow<NetworkDiagnosticResult?>(null)
    val diagnosticResult: StateFlow<NetworkDiagnosticResult?> = _diagnosticResult.asStateFlow()

    private val _dataUsage = MutableStateFlow(AppDataUsage())
    val dataUsage: StateFlow<AppDataUsage> = _dataUsage.asStateFlow()

    private val _networkEvents = MutableSharedFlow<NetworkEvent>(
        replay = 0,
        extraBufferCapacity = 100,
        onBufferOverflow = kotlinx.coroutines.channels.BufferOverflow.DROP_OLDEST
    )
    val networkEvents: SharedFlow<NetworkEvent> = _networkEvents.asSharedFlow()

    private var connectivityNetworkCallback: ConnectivityNetworkCallback? = null
    private var telephonyCallback: TelephonyCallback? = null
    private var wifiBroadcastReceiver: WifiBroadcastReceiver? = null
    private var batteryReceiver: BatteryBroadcastReceiver? = null

    private val isRegistered = AtomicBoolean(false)
    private val isDiagnosticRunning = AtomicBoolean(false)
    private val isCaptivePortalCheckRunning = AtomicBoolean(false)

    private val lastCaptivePortalCheckTime = AtomicLong(0L)
    private val lastBandwidthEstimateTime = AtomicLong(0L)
    private val lastDiagnosticTime = AtomicLong(0L)
    private val lastDataUsageCheckTime = AtomicLong(0L)

    private val activeNetworks = ConcurrentHashMap<Long, NetworkMetadata>()
    private val latencyHistory = ConcurrentLinkedDeque<LatencySample>()
    private val bandwidthSamples = ConcurrentLinkedDeque<BandwidthSample>()
    private val packetLossHistory = ConcurrentLinkedDeque<PacketLossSample>()
    private val networkSwitchHistory = ConcurrentLinkedDeque<NetworkSwitchEvent>()
    private val diagnosticHistory = ConcurrentLinkedDeque<NetworkDiagnosticResult>()

    private val totalBytesDownloaded = AtomicLong(0L)
    private val totalBytesUploaded = AtomicLong(0L)
    private val sessionBytesDownloaded = AtomicLong(0L)
    private val sessionBytesUploaded = AtomicLong(0L)

    private val currentNetworkHandle = AtomicLong(-1L)
    private val lastNetworkSwitchTime = AtomicLong(System.currentTimeMillis())
    private val uptimeSeconds = AtomicLong(0L)

    private val trustAllCertificates = arrayOf<TrustManager>(object : X509TrustManager {
        override fun checkClientTrusted(chain: Array<java.security.cert.X509Certificate>, authType: String) {}
        override fun checkServerTrusted(chain: Array<java.security.cert.X509Certificate>, authType: String) {}
        override fun getAcceptedIssuers(): Array<java.security.cert.X509Certificate> = arrayOf()
    })

    data class ComprehensiveNetworkState(
        val isConnected: Boolean = false,
        val networkType: NetworkType = NetworkType.UNKNOWN,
        val isMetered: Boolean = false,
        val isCaptivePortal: Boolean = false,
        val isVpnActive: Boolean = false,
        val isRoaming: Boolean = false,
        val isDataSaverActive: Boolean = false,
        val isBatterySaverActive: Boolean = false,
        val isPowerSaveMode: Boolean = false,
        val signalStrength: Int = -1,
        val signalStrengthDbm: Int = Int.MIN_VALUE,
        val linkSpeedMbps: Int = 0,
        val frequencyMhz: Int = 0,
        val wifiStandard: WifiStandard = WifiStandard.UNKNOWN,
        val wifiSsid: String? = null,
        val wifiBssid: String? = null,
        val wifiRssi: Int = Int.MIN_VALUE,
        val mobileNetworkType: MobileNetworkType = MobileNetworkType.UNKNOWN,
        val mobileOperator: String? = null,
        val mobileMcc: Int = 0,
        val mobileMnc: Int = 0,
        val dnsServers: List<String> = emptyList(),
        val gateway: String? = null,
        val interfaceName: String? = null,
        val ipAddresses: List<String> = emptyList(),
        val mtu: Int = 0,
        val lastChecked: Long = 0L,
        val uptimeSeconds: Long = 0L,
        val totalDownloadedBytes: Long = 0L,
        val totalUploadedBytes: Long = 0L,
        val sessionDownloadedBytes: Long = 0L,
        val sessionUploadedBytes: Long = 0L
    )

    data class NetworkMetadata(
        val network: Network,
        val capabilities: NetworkCapabilities,
        val linkProperties: LinkProperties,
        val connectedAt: Long = System.currentTimeMillis(),
        val isActive: Boolean = true
    )

    data class LatencySample(
        val timestamp: Long,
        val target: String,
        val latencyMs: Long,
        val success: Boolean,
        val error: String? = null
    )

    data class BandwidthSample(
        val timestamp: Long,
        val bytesTransferred: Long,
        val durationMs: Long,
        val type: SampleType,
        val success: Boolean,
        val error: String? = null
    )

    data class PacketLossSample(
        val timestamp: Long,
        val target: String,
        val sentPackets: Int,
        val receivedPackets: Int,
        val lossPercent: Float
    )

    data class NetworkSwitchEvent(
        val timestamp: Long,
        val fromType: NetworkType,
        val toType: NetworkType,
        val reason: String
    )

    data class NetworkDiagnosticResult(
        val timestamp: Long = System.currentTimeMillis(),
        val isConnected: Boolean = false,
        val dnsResolution: DnsDiagnostic = DnsDiagnostic(),
        val tcpConnectivity: TcpDiagnostic = TcpDiagnostic(),
        val httpProbe: HttpDiagnostic = HttpDiagnostic(),
        val imapProbe: ProtocolDiagnostic = ProtocolDiagnostic(),
        val smtpProbe: ProtocolDiagnostic = ProtocolDiagnostic(),
        val overallHealth: DiagnosticHealth = DiagnosticHealth.UNKNOWN,
        val issues: List<String> = emptyList(),
        val recommendations: List<String> = emptyList()
    )

    data class DnsDiagnostic(
        val success: Boolean = false,
        val hostname: String = "",
        val resolvedAddresses: List<String> = emptyList(),
        val resolutionTimeMs: Long = 0L,
        val error: String? = null
    )

    data class TcpDiagnostic(
        val success: Boolean = false,
        val host: String = "",
        val port: Int = 0,
        val connectTimeMs: Long = 0L,
        val tlsHandshakeTimeMs: Long = 0L,
        val tlsVersion: String? = null,
        val error: String? = null
    )

    data class HttpDiagnostic(
        val success: Boolean = false,
        val url: String = "",
        val responseCode: Int = 0,
        val responseTimeMs: Long = 0L,
        val redirectUrl: String? = null,
        val isCaptivePortal: Boolean = false,
        val error: String? = null
    )

    data class ProtocolDiagnostic(
        val success: Boolean = false,
        val host: String = "",
        val port: Int = 0,
        val connectTimeMs: Long = 0L,
        val greetingReceived: Boolean = false,
        val greetingText: String? = null,
        val error: String? = null
    )

    data class AppDataUsage(
        val wifiDownloadBytes: Long = 0L,
        val wifiUploadBytes: Long = 0L,
        val mobileDownloadBytes: Long = 0L,
        val mobileUploadBytes: Long = 0L,
        val lastResetTimestamp: Long = 0L
    )

    data class BandwidthEstimate(
        val downloadKbps: Long = 0L,
        val uploadKbps: Long = 0L,
        val latencyMs: Long = 0L,
        val jitterMs: Long = 0L,
        val packetLossPercent: Float = 0f,
        val sampleCount: Int = 0,
        val confidence: Float = 0f,
        val lastUpdated: Long = 0L
    ) {
        companion object {
            val UNKNOWN = BandwidthEstimate()
        }
    }

    enum class SampleType { DOWNLOAD, UPLOAD, LATENCY, PACKET_LOSS }

    enum class ConnectionQuality(val minBandwidthKbps: Long, val maxLatencyMs: Long) {
        EXCELLENT(5000, 50),
        GOOD(2000, 150),
        FAIR(500, 400),
        POOR(100, 1000),
        UNKNOWN(0, Long.MAX_VALUE),
        OFFLINE(0, Long.MAX_VALUE)
    }

    enum class WifiStandard {
        UNKNOWN, WIFI_4, WIFI_5, WIFI_6, WIFI_6E, WIFI_7
    }

    enum class MobileNetworkType {
        UNKNOWN, GPRS, EDGE, UMTS, HSDPA, HSPA, HSPAP, LTE, LTE_CA, NR, NR_SA
    }

    enum class DiagnosticHealth {
        EXCELLENT, GOOD, FAIR, POOR, CRITICAL, UNKNOWN
    }

    sealed class NetworkEvent {
        data class Connected(val type: NetworkType) : NetworkEvent()
        data class Disconnected(val reason: String) : NetworkEvent()
        data class Switched(val from: NetworkType, val to: NetworkType) : NetworkEvent()
        data class QualityChanged(val old: ConnectionQuality, val new: ConnectionQuality) : NetworkEvent()
        data class CaptivePortalDetected(val network: String) : NetworkEvent()
        data class CaptivePortalCleared(val network: String) : NetworkEvent()
        data class BandwidthUpdated(val estimate: BandwidthEstimate) : NetworkEvent()
        data class DiagnosticCompleted(val result: NetworkDiagnosticResult) : NetworkEvent()
        data object RoamingStarted : NetworkEvent()
        data object RoamingEnded : NetworkEvent()
        data class VpnConnected(val interfaceName: String) : NetworkEvent()
        data class VpnDisconnected(val interfaceName: String) : NetworkEvent()
        data class DataSaverToggled(val enabled: Boolean) : NetworkEvent()
        data class BatterySaverToggled(val enabled: Boolean) : NetworkEvent()
    }

    init {
        registerCallbacks()
        registerReceivers()
    }

    fun startMonitoring() {
        if (isRegistered.compareAndSet(false, true)) {
            logger.i(TAG, "Starting network monitoring")
            scope.launch {
                performInitialSetup()
                startPeriodicTasks()
            }
        }
    }

    fun stopMonitoring() {
        if (isRegistered.compareAndSet(true, false)) {
            logger.i(TAG, "Stopping network monitoring")
            unregisterCallbacks()
            unregisterReceivers()
        }
    }

    fun forceRefresh() {
        scope.launch {
            performFullRefresh()
        }
    }

    suspend fun runFullDiagnostics(): NetworkDiagnosticResult {
        if (!isDiagnosticRunning.compareAndSet(false, true)) {
            return _diagnosticResult.value ?: NetworkDiagnosticResult()
        }

        try {
            logger.i(TAG, "Starting full network diagnostics")

            val dns = testDnsResolution()
            val tcp = testTcpConnectivity()
            val http = testHttpProbe()
            val imap = testProtocolConnectivity("imap.gmail.com", 993, "IMAP")
            val smtp = testProtocolConnectivity("smtp.gmail.com", 587, "SMTP")

            val issues = mutableListOf<String>()
            val recommendations = mutableListOf<String>()

            if (!dns.success) {
                issues.add("DNS resolution failed")
                recommendations.add("Check DNS settings or try using 8.8.8.8")
            }
            if (!tcp.success) {
                issues.add("TCP connectivity failed")
                recommendations.add("Check firewall or proxy settings")
            }
            if (http.isCaptivePortal) {
                issues.add("Captive portal detected")
                recommendations.add("Complete portal authentication")
            }
            if (!imap.success) {
                issues.add("IMAP connectivity blocked")
                recommendations.add("Check if IMAP ports are allowed on network")
            }
            if (!smtp.success) {
                issues.add("SMTP connectivity blocked")
                recommendations.add("Check if SMTP ports are allowed on network")
            }

            val health = when {
                dns.success && tcp.success && !http.isCaptivePortal && imap.success -> DiagnosticHealth.EXCELLENT
                dns.success && tcp.success && !http.isCaptivePortal -> DiagnosticHealth.GOOD
                dns.success && tcp.success -> DiagnosticHealth.FAIR
                dns.success -> DiagnosticHealth.POOR
                else -> DiagnosticHealth.CRITICAL
            }

            val result = NetworkDiagnosticResult(
                isConnected = _isOnline.value,
                dnsResolution = dns,
                tcpConnectivity = tcp,
                httpProbe = http,
                imapProbe = imap,
                smtpProbe = smtp,
                overallHealth = health,
                issues = issues,
                recommendations = recommendations
            )

            _diagnosticResult.value = result
            diagnosticHistory.addLast(result)
            if (diagnosticHistory.size > 20) diagnosticHistory.removeFirst()

            lastDiagnosticTime.set(System.currentTimeMillis())
            _networkEvents.tryEmit(NetworkEvent.DiagnosticCompleted(result))

            logger.i(TAG, "Diagnostics completed: health=$health, issues=${issues.size}")
            return result

        } catch (e: Exception) {
            logger.e(TAG, "Diagnostics failed", e)
            return NetworkDiagnosticResult(overallHealth = DiagnosticHealth.UNKNOWN)
        } finally {
            isDiagnosticRunning.set(false)
        }
    }

    fun isNetworkAvailable(): Boolean {
        return _isOnline.value && !_networkState.value.isCaptivePortal
    }

    fun isWifiConnected(): Boolean {
        return _networkType.value == NetworkType.WIFI
    }

    fun isMobileDataConnected(): Boolean {
        return _networkType.value in listOf(
            NetworkType.MOBILE, NetworkType.MOBILE_2G, NetworkType.MOBILE_3G,
            NetworkType.MOBILE_4G, NetworkType.MOBILE_5G
        )
    }

    fun isEthernetConnected(): Boolean {
        return _networkType.value == NetworkType.ETHERNET
    }

    fun isVpnActive(): Boolean {
        return _networkState.value.isVpnActive
    }

    fun isMeteredConnection(): Boolean {
        return _networkState.value.isMetered
    }

    fun isRoaming(): Boolean {
        return _networkState.value.isRoaming
    }

    fun shouldDeferLargeDownloads(): Boolean {
        return isMeteredConnection() ||
                _networkState.value.isDataSaverActive ||
                _connectionQuality.value == ConnectionQuality.POOR ||
                _bandwidthEstimate.value.downloadKbps < ConnectionQuality.FAIR.minBandwidthKbps
    }

    fun shouldPauseSync(): Boolean {
        return !_isOnline.value ||
                _networkState.value.isBatterySaverActive ||
                _connectionQuality.value == ConnectionQuality.POOR ||
                (_networkType.value == NetworkType.MOBILE_2G) ||
                (_bandwidthEstimate.value.downloadKbps < 50 && _networkType.value.isMobile())
    }

    fun getOptimalSyncStrategy(): SyncStrategy {
        return when {
            !_isOnline.value -> SyncStrategy.OFFLINE
            _networkState.value.isBatterySaverActive -> SyncStrategy.MINIMAL
            _networkType.value == NetworkType.WIFI -> SyncStrategy.UNMETERED
            _connectionQuality.value == ConnectionQuality.EXCELLENT -> SyncStrategy.AGGRESSIVE
            _connectionQuality.value == ConnectionQuality.GOOD -> SyncStrategy.NORMAL
            _connectionQuality.value == ConnectionQuality.FAIR -> SyncStrategy.CONSERVATIVE
            _connectionQuality.value == ConnectionQuality.POOR -> SyncStrategy.MINIMAL
            _networkState.value.isRoaming -> SyncStrategy.MINIMAL
            else -> SyncStrategy.CONSERVATIVE
        }
    }

    enum class SyncStrategy {
        AGGRESSIVE, NORMAL, CONSERVATIVE, MINIMAL, UNMETERED, OFFLINE
    }

    fun getSignalStrengthPercent(): Int {
        return when (_networkType.value) {
            NetworkType.WIFI -> {
                val rssi = _networkState.value.wifiRssi
                if (rssi == Int.MIN_VALUE) -1
                else ((rssi + 100) * 2).coerceIn(0, 100)
            }
            NetworkType.MOBILE, NetworkType.MOBILE_2G, NetworkType.MOBILE_3G,
            NetworkType.MOBILE_4G, NetworkType.MOBILE_5G -> {
                val level = _networkState.value.signalStrength
                if (level < 0) -1 else level * 25
            }
            else -> -1
        }
    }

    fun getConnectionSummary(): String {
        val state = _networkState.value
        val quality = _connectionQuality.value
        val bandwidth = _bandwidthEstimate.value

        return buildString {
            append(if (state.isConnected) "Connected" else "Disconnected")
            if (state.isConnected) {
                append(" via ${state.networkType.displayName}")
                append(" (${quality.name.lowercase()} quality)")
                if (bandwidth.downloadKbps > 0) {
                    append(", ${bandwidth.downloadKbps}kbps")
                }
                if (state.isMetered) append(", metered")
                if (state.isVpnActive) append(", VPN")
                if (state.isCaptivePortal) append(", captive portal")
            }
        }
    }

    private suspend fun performInitialSetup() {
        val state = buildComprehensiveNetworkState()
        updateAllStates(state)
        checkDataUsage()

        if (state.isConnected) {
            checkCaptivePortal()
            estimateBandwidth()
            evaluateConnectionQuality()
        }
    }

    private suspend fun startPeriodicTasks() {
        while (isActive && isRegistered.get()) {
            try {
                performFullRefresh()
                delay(30_000L)
            } catch (e: CancellationException) {
                break
            } catch (e: Exception) {
                logger.w(TAG, "Periodic task failed, retrying in 10s", e)
                delay(10_000L)
            }
        }
    }

    private suspend fun performFullRefresh() {
        val previousType = _networkType.value
        val state = buildComprehensiveNetworkState()
        updateAllStates(state)

        if (state.isConnected) {
            if (state.networkType != previousType) {
                handleNetworkSwitch(previousType, state.networkType)
            }

            val latency = measureLatencyToMultipleTargets()
            updateLatencyHistory(latency)

            if (shouldCheckCaptivePortal()) {
                checkCaptivePortal()
            }

            if (shouldEstimateBandwidth()) {
                estimateBandwidth()
            }

            if (shouldRunDiagnostics()) {
                runFullDiagnostics()
            }

            if (shouldCheckDataUsage()) {
                checkDataUsage()
            }

            evaluateConnectionQuality()
            updateUptime()
            trackBytesTransferred()
        } else {
            _connectionQuality.value = ConnectionQuality.OFFLINE
            _bandwidthEstimate.value = BandwidthEstimate.UNKNOWN
        }
    }

    @SuppressLint("MissingPermission")
    private fun buildComprehensiveNetworkState(): ComprehensiveNetworkState {
        val cm = connectivityManager ?: return ComprehensiveNetworkState()
        val activeNetwork = cm.activeNetwork ?: return ComprehensiveNetworkState()
        val capabilities = cm.getNetworkCapabilities(activeNetwork)
            ?: return ComprehensiveNetworkState()
        val linkProperties = cm.getLinkProperties(activeNetwork)

        val networkType = determineNetworkType(capabilities)
        val isMetered = cm.isActiveNetworkMetered
        val isVpn = capabilities.hasTransport(NetworkCapabilities.TRANSPORT_VPN)
        val isDataSaver = isDataSaverEnabled()
        val isBatterySaver = isBatterySaverEnabled()
        val isPowerSave = isPowerSaveModeEnabled()

        val signalInfo = getSignalInfo(networkType)
        val wifiInfo = getWifiInfo()
        val mobileInfo = getMobileInfo()
        val networkInterfaceInfo = getNetworkInterfaceInfo(linkProperties)

        val totalDl = totalBytesDownloaded.get()
        val totalUl = totalBytesUploaded.get()
        val sessionDl = sessionBytesDownloaded.get()
        val sessionUl = sessionBytesUploaded.get()

        return ComprehensiveNetworkState(
            isConnected = true,
            networkType = networkType,
            isMetered = isMetered,
            isVpnActive = isVpn,
            isRoaming = mobileInfo.isRoaming,
            isDataSaverActive = isDataSaver,
            isBatterySaverActive = isBatterySaver,
            isPowerSaveMode = isPowerSave,
            signalStrength = signalInfo.level,
            signalStrengthDbm = signalInfo.dbm,
            linkSpeedMbps = capabilities.linkDownstreamBandwidthKbps / 1000,
            frequencyMhz = wifiInfo.frequencyMhz,
            wifiStandard = wifiInfo.standard,
            wifiSsid = wifiInfo.ssid,
            wifiBssid = wifiInfo.bssid,
            wifiRssi = wifiInfo.rssi,
            mobileNetworkType = mobileInfo.networkType,
            mobileOperator = mobileInfo.operator,
            mobileMcc = mobileInfo.mcc,
            mobileMnc = mobileInfo.mnc,
            dnsServers = networkInterfaceInfo.dnsServers,
            gateway = networkInterfaceInfo.gateway,
            interfaceName = networkInterfaceInfo.interfaceName,
            ipAddresses = networkInterfaceInfo.ipAddresses,
            mtu = networkInterfaceInfo.mtu,
            lastChecked = System.currentTimeMillis(),
            uptimeSeconds = uptimeSeconds.get(),
            totalDownloadedBytes = totalDl,
            totalUploadedBytes = totalUl,
            sessionDownloadedBytes = sessionDl,
            sessionUploadedBytes = sessionUl
        )
    }

    private data class SignalInfo(val level: Int, val dbm: Int)
    private data class WifiInfoData(
        val ssid: String?, val bssid: String?, val rssi: Int,
        val frequencyMhz: Int, val standard: WifiStandard
    )
    private data class MobileInfoData(
        val networkType: MobileNetworkType, val operator: String?,
        val mcc: Int, val mnc: Int, val isRoaming: Boolean
    )
    private data class NetworkInterfaceData(
        val dnsServers: List<String>, val gateway: String?,
        val interfaceName: String?, val ipAddresses: List<String>, val mtu: Int
    )

    private fun getSignalInfo(networkType: NetworkType): SignalInfo {
        return when (networkType) {
            NetworkType.WIFI -> {
                val rssi = wifiManager?.connectionInfo?.rssi ?: Int.MIN_VALUE
                val level = if (rssi != Int.MIN_VALUE) {
                    WifiManager.calculateSignalLevel(rssi, 5)
                } else -1
                SignalInfo(level, rssi)
            }
            else -> {
                try {
                    val tm = telephonyManager ?: return SignalInfo(-1, Int.MIN_VALUE)
                    val strength = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                        tm.signalStrength
                    } else {
                        @Suppress("DEPRECATION") tm.signalStrength
                    }
                    SignalInfo(strength?.level ?: -1, strength?.dbm ?: Int.MIN_VALUE)
                } catch (e: SecurityException) {
                    SignalInfo(-1, Int.MIN_VALUE)
                }
            }
        }
    }

    @SuppressLint("MissingPermission")
    private fun getWifiInfo(): WifiInfoData {
        val wm = wifiManager ?: return WifiInfoData(null, null, Int.MIN_VALUE, 0, WifiStandard.UNKNOWN)
        return try {
            val info = wm.connectionInfo
            val standard = when {
                Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU &&
                info.wifiStandard == ScanResult.WIFI_STANDARD_11BE -> WifiStandard.WIFI_7
                Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU &&
                info.wifiStandard == ScanResult.WIFI_STANDARD_11AX -> WifiStandard.WIFI_6
                Build.VERSION.SDK_INT >= Build.VERSION_CODES.R &&
                info.wifiStandard == ScanResult.WIFI_STANDARD_11AC -> WifiStandard.WIFI_5
                info.frequency in 5000..6000 -> WifiStandard.WIFI_5
                info.frequency in 2400..2500 -> WifiStandard.WIFI_4
                else -> WifiStandard.UNKNOWN
            }
            WifiInfoData(
                ssid = info.ssid?.removeSurrounding("\""),
                bssid = info.bssid,
                rssi = info.rssi,
                frequencyMhz = info.frequency,
                standard = standard
            )
        } catch (e: SecurityException) {
            WifiInfoData(null, null, Int.MIN_VALUE, 0, WifiStandard.UNKNOWN)
        }
    }

    private fun getMobileInfo(): MobileInfoData {
        val tm = telephonyManager ?: return MobileInfoData(MobileNetworkType.UNKNOWN, null, 0, 0, false)
        return try {
            val networkType = when {
                Build.VERSION.SDK_INT >= Build.VERSION_CODES.R -> {
                    when (tm.dataNetworkType) {
                        TelephonyManager.NETWORK_TYPE_NR -> MobileNetworkType.NR
                        TelephonyManager.NETWORK_TYPE_LTE -> MobileNetworkType.LTE
                        TelephonyManager.NETWORK_TYPE_LTE_CA -> MobileNetworkType.LTE_CA
                        TelephonyManager.NETWORK_TYPE_HSPAP -> MobileNetworkType.HSPAP
                        TelephonyManager.NETWORK_TYPE_HSDPA -> MobileNetworkType.HSDPA
                        TelephonyManager.NETWORK_TYPE_HSUPA -> MobileNetworkType.HSPA
                        TelephonyManager.NETWORK_TYPE_HSPA -> MobileNetworkType.HSPA
                        TelephonyManager.NETWORK_TYPE_UMTS -> MobileNetworkType.UMTS
                        TelephonyManager.NETWORK_TYPE_EDGE -> MobileNetworkType.EDGE
                        TelephonyManager.NETWORK_TYPE_GPRS -> MobileNetworkType.GPRS
                        else -> MobileNetworkType.UNKNOWN
                    }
                }
                else -> MobileNetworkType.UNKNOWN
            }
            MobileInfoData(
                networkType = networkType,
                operator = tm.networkOperatorName,
                mcc = try { tm.networkOperator.substring(0, 3).toInt() } catch (_: Exception) { 0 },
                mnc = try { tm.networkOperator.substring(3).toInt() } catch (_: Exception) { 0 },
                isRoaming = tm.isNetworkRoaming
            )
        } catch (e: SecurityException) {
            MobileInfoData(MobileNetworkType.UNKNOWN, null, 0, 0, false)
        }
    }

    private fun getNetworkInterfaceInfo(linkProperties: LinkProperties?): NetworkInterfaceData {
        if (linkProperties == null) return NetworkInterfaceData(emptyList(), null, null, emptyList(), 0)
        return NetworkInterfaceData(
            dnsServers = linkProperties.dnsServers.map { it.hostAddress ?: "unknown" },
            gateway = linkProperties.routes.firstOrNull()?.gateway?.hostAddress,
            interfaceName = linkProperties.interfaceName,
            ipAddresses = linkProperties.linkAddresses.map { it.address.hostAddress ?: "unknown" },
            mtu = linkProperties.mtu
        )
    }

    private fun determineNetworkType(capabilities: NetworkCapabilities): NetworkType {
        return when {
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_WIFI) -> NetworkType.WIFI
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_ETHERNET) -> NetworkType.ETHERNET
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_VPN) -> NetworkType.VPN
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR) -> determineCellularType()
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_BLUETOOTH) -> NetworkType.BLUETOOTH
            else -> NetworkType.UNKNOWN
        }
    }

    private fun determineCellularType(): NetworkType {
        val tm = telephonyManager ?: return NetworkType.MOBILE
        return try {
            when (tm.dataNetworkType) {
                TelephonyManager.NETWORK_TYPE_NR -> NetworkType.MOBILE_5G
                TelephonyManager.NETWORK_TYPE_LTE, TelephonyManager.NETWORK_TYPE_LTE_CA -> NetworkType.MOBILE_4G
                TelephonyManager.NETWORK_TYPE_EVDO_0, TelephonyManager.NETWORK_TYPE_EVDO_A,
                TelephonyManager.NETWORK_TYPE_EVDO_B, TelephonyManager.NETWORK_TYPE_EHRPD,
                TelephonyManager.NETWORK_TYPE_HSPA, TelephonyManager.NETWORK_TYPE_HSPAP,
                TelephonyManager.NETWORK_TYPE_HSDPA, TelephonyManager.NETWORK_TYPE_HSUPA,
                TelephonyManager.NETWORK_TYPE_UMTS -> NetworkType.MOBILE_3G
                TelephonyManager.NETWORK_TYPE_EDGE, TelephonyManager.NETWORK_TYPE_GPRS -> NetworkType.MOBILE_2G
                else -> NetworkType.MOBILE
            }
        } catch (e: SecurityException) {
            NetworkType.MOBILE
        }
    }

    private fun isDataSaverEnabled(): Boolean {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
            connectivityManager?.restrictBackgroundStatus == ConnectivityManager.RESTRICT_BACKGROUND_STATUS_ENABLED
        } else false
    }

    @SuppressLint("NewApi")
    private fun isBatterySaverEnabled(): Boolean {
        return powerManager?.isPowerSaveMode ?: false
    }

    @SuppressLint("NewApi")
    private fun isPowerSaveModeEnabled(): Boolean {
        return powerManager?.isPowerSaveMode ?: false
    }

    private suspend fun measureLatencyToMultipleTargets(): List<LatencySample> {
        val targets = listOf(
            Triple("8.8.8.8", 53, "Google DNS"),
            Triple("1.1.1.1", 53, "Cloudflare DNS"),
            Triple("208.67.222.222", 53, "OpenDNS"),
            Triple("www.google.com", 443, "Google HTTPS"),
            Triple("imap.gmail.com", 993, "Gmail IMAP")
        )

        return targets.map { (host, port, label) ->
            async {
                try {
                    val socket = Socket()
                    val address = InetSocketAddress(host, port)
                    val startTime = System.nanoTime()
                    withTimeout(5000L) { socket.connect(address, 5000) }
                    val latency = (System.nanoTime() - startTime) / 1_000_000
                    socket.close()
                    LatencySample(System.currentTimeMillis(), label, latency, true)
                } catch (e: Exception) {
                    LatencySample(System.currentTimeMillis(), label, -1L, false, e.message)
                }
            }
        }.awaitAll()
    }

    private fun updateLatencyHistory(samples: List<LatencySample>) {
        samples.forEach { sample ->
            latencyHistory.addLast(sample)
            if (latencyHistory.size > 100) latencyHistory.removeFirst()
        }
    }

    private fun getAverageLatency(): Long {
        val recentSamples = latencyHistory.takeLast(20).filter { it.success }
        if (recentSamples.isEmpty()) return -1L
        val sorted = recentSamples.map { it.latencyMs }.sorted()
        val trimmed = sorted.subList(sorted.size / 4, sorted.size * 3 / 4)
        return if (trimmed.isNotEmpty()) trimmed.average().toLong() else -1L
    }

    private fun getLatencyJitter(): Long {
        val recentSamples = latencyHistory.takeLast(20).filter { it.success }
        if (recentSamples.size < 3) return 0L
        val latencies = recentSamples.map { it.latencyMs }
        var totalVariation = 0L
        for (i in 1 until latencies.size) {
            totalVariation += abs(latencies[i] - latencies[i - 1])
        }
        return totalVariation / (latencies.size - 1)
    }

    private suspend fun testDnsResolution(): DnsDiagnostic {
        val hostname = "imap.gmail.com"
        return try {
            val startTime = System.nanoTime()
            val addresses = withTimeout(5000L) {
                InetAddress.getAllByName(hostname).map { it.hostAddress ?: "unknown" }
            }
            val resolutionTime = (System.nanoTime() - startTime) / 1_000_000
            DnsDiagnostic(true, hostname, addresses, resolutionTime)
        } catch (e: Exception) {
            DnsDiagnostic(false, hostname, error = e.message)
        }
    }

    private suspend fun testTcpConnectivity(): TcpDiagnostic {
        return try {
            val socket = SSLSocketFactory.getDefault().createSocket() as SSLSocket
            val address = InetSocketAddress("imap.gmail.com", 993)

            val connectStart = System.nanoTime()
            withTimeout(10000L) { socket.connect(address, 10000) }
            val connectTime = (System.nanoTime() - connectStart) / 1_000_000

            val tlsStart = System.nanoTime()
            socket.startHandshake()
            val tlsTime = (System.nanoTime() - tlsStart) / 1_000_000

            val session = socket.session
            socket.close()

            TcpDiagnostic(
                success = true,
                host = "imap.gmail.com",
                port = 993,
                connectTimeMs = connectTime,
                tlsHandshakeTimeMs = tlsTime,
                tlsVersion = session.protocol
            )
        } catch (e: Exception) {
            TcpDiagnostic(
                success = false,
                host = "imap.gmail.com",
                port = 993,
                error = e.message
            )
        }
    }

    private suspend fun testHttpProbe(): HttpDiagnostic {
        val urls = listOf(
            "http://clients3.google.com/generate_204",
            "http://connectivitycheck.gstatic.com/generate_204"
        )

        for (url in urls) {
            try {
                val connection = withTimeout(10000L) {
                    URL(url).openConnection() as HttpURLConnection
                }
                connection.apply {
                    connectTimeout = 5000
                    readTimeout = 5000
                    instanceFollowRedirects = false
                    setRequestProperty("User-Agent", "MailX-NetworkProbe/1.0")
                }

                val startTime = System.nanoTime()
                val responseCode = connection.responseCode
                val responseTime = (System.nanoTime() - startTime) / 1_000_000

                val isRedirect = responseCode in 300..399
                val redirectUrl = if (isRedirect) connection.getHeaderField("Location") else null

                connection.disconnect()

                return HttpDiagnostic(
                    success = responseCode in listOf(204, 200),
                    url = url,
                    responseCode = responseCode,
                    responseTimeMs = responseTime,
                    redirectUrl = redirectUrl,
                    isCaptivePortal = isRedirect
                )
            } catch (_: Exception) {
                continue
            }
        }

        return HttpDiagnostic(success = false, error = "All HTTP probes failed")
    }

    private suspend fun testProtocolConnectivity(
        host: String,
        port: Int,
        protocol: String
    ): ProtocolDiagnostic {
        return try {
            val socket = Socket()
            val address = InetSocketAddress(host, port)

            val connectStart = System.nanoTime()
            withTimeout(10000L) { socket.connect(address, 10000) }
            val connectTime = (System.nanoTime() - connectStart) / 1_000_000

            socket.soTimeout = 5000
            val reader = BufferedReader(InputStreamReader(socket.getInputStream()))
            val greeting = withTimeout(5000L) { reader.readLine() }

            socket.close()

            ProtocolDiagnostic(
                success = greeting != null,
                host = host,
                port = port,
                connectTimeMs = connectTime,
                greetingReceived = greeting != null,
                greetingText = greeting
            )
        } catch (e: Exception) {
            ProtocolDiagnostic(
                success = false,
                host = host,
                port = port,
                error = e.message
            )
        }
    }

    private suspend fun checkCaptivePortal() {
        if (!isCaptivePortalCheckRunning.compareAndSet(false, true)) return

        try {
            val httpProbe = testHttpProbe()
            val isPortal = httpProbe.isCaptivePortal

            if (isPortal != _networkState.value.isCaptivePortal) {
                _networkState.update { it.copy(isCaptivePortal = isPortal) }
                _isOnline.update { !isPortal }

                if (isPortal) {
                    _networkEvents.tryEmit(NetworkEvent.CaptivePortalDetected(
                        wifiManager?.connectionInfo?.ssid ?: "unknown"
                    ))
                } else {
                    _networkEvents.tryEmit(NetworkEvent.CaptivePortalCleared(
                        wifiManager?.connectionInfo?.ssid ?: "unknown"
                    ))
                }
            }

            lastCaptivePortalCheckTime.set(System.currentTimeMillis())
        } catch (e: Exception) {
            logger.w(TAG, "Captive portal check failed", e)
        } finally {
            isCaptivePortalCheckRunning.set(false)
        }
    }

    private suspend fun estimateBandwidth() {
        try {
            val sample = performBandwidthSample()
            if (sample != null) {
                bandwidthSamples.addLast(sample)
                if (bandwidthSamples.size > 20) bandwidthSamples.removeFirst()

                val estimate = calculateBandwidthEstimate()
                _bandwidthEstimate.value = estimate
                lastBandwidthEstimateTime.set(System.currentTimeMillis())

                _networkEvents.tryEmit(NetworkEvent.BandwidthUpdated(estimate))

                logger.d(TAG, "Bandwidth: ${estimate.downloadKbps}kbps, " +
                        "latency: ${estimate.latencyMs}ms, jitter: ${estimate.jitterMs}ms")
            }
        } catch (e: Exception) {
            logger.w(TAG, "Bandwidth estimation failed", e)
        }
    }

    private suspend fun performBandwidthSample(): BandwidthSample? {
        val testUrls = listOf(
            "https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png",
            "https://www.gstatic.com/webp/gallery/1.webp",
            "https://www.google.com/favicon.ico"
        )

        for (url in testUrls) {
            try {
                return withTimeout(15000L) {
                    withContext(Dispatchers.IO) {
                        val connection = URL(url).openConnection() as HttpURLConnection
                        connection.apply {
                            connectTimeout = 5000
                            readTimeout = 10000
                            setRequestProperty("User-Agent", "MailX-BandwidthTest/1.0")
                            setRequestProperty("Cache-Control", "no-cache")
                        }

                        val startTime = System.nanoTime()
                        val bytes = connection.inputStream.readBytes()
                        val durationNs = System.nanoTime() - startTime
                        connection.disconnect()

                        val durationMs = durationNs / 1_000_000
                        val bytesTransferred = bytes.size.toLong()

                        if (durationMs > 0 && bytesTransferred > 500) {
                            BandwidthSample(
                                timestamp = System.currentTimeMillis(),
                                bytesTransferred = bytesTransferred,
                                durationMs = durationMs,
                                type = SampleType.DOWNLOAD,
                                success = true
                            )
                        } else null
                    }
                }
            } catch (_: Exception) { continue }
        }
        return null
    }

    private fun calculateBandwidthEstimate(): BandwidthEstimate {
        val samples = bandwidthSamples.takeLast(10).filter { it.success }
        if (samples.isEmpty()) return BandwidthEstimate.UNKNOWN

        val totalBytes = samples.sumOf { it.bytesTransferred }
        val totalDurationMs = samples.sumOf { it.durationMs }
        val downloadKbps = if (totalDurationMs > 0) {
            (totalBytes * 8L * 1000L) / totalDurationMs / 1000L
        } else 0L

        val avgLatency = getAverageLatency()
        val jitter = getLatencyJitter()
        val packetLoss = calculatePacketLoss()
        val confidence = when {
            samples.size >= 8 -> 0.9f
            samples.size >= 5 -> 0.7f
            samples.size >= 3 -> 0.5f
            else -> 0.3f
        }

        return BandwidthEstimate(
            downloadKbps = downloadKbps,
            uploadKbps = (downloadKbps * 0.6).toLong(),
            latencyMs = avgLatency,
            jitterMs = jitter,
            packetLossPercent = packetLoss,
            sampleCount = samples.size,
            confidence = confidence,
            lastUpdated = System.currentTimeMillis()
        )
    }

    private fun calculatePacketLoss(): Float {
        val samples = packetLossHistory.takeLast(10)
        if (samples.isEmpty()) return 0f
        return samples.map { it.lossPercent }.average().toFloat()
    }

    private fun evaluateConnectionQuality() {
        val bandwidth = _bandwidthEstimate.value
        val latency = getAverageLatency()
        val packetLoss = calculatePacketLoss()

        val quality = when {
            !_isOnline.value -> ConnectionQuality.OFFLINE
            bandwidth.downloadKbps >= ConnectionQuality.EXCELLENT.minBandwidthKbps &&
            latency in 1..ConnectionQuality.EXCELLENT.maxLatencyMs &&
            packetLoss < 2f -> ConnectionQuality.EXCELLENT
            bandwidth.downloadKbps >= ConnectionQuality.GOOD.minBandwidthKbps &&
            latency in 1..ConnectionQuality.GOOD.maxLatencyMs &&
            packetLoss < 5f -> ConnectionQuality.GOOD
            bandwidth.downloadKbps >= ConnectionQuality.FAIR.minBandwidthKbps &&
            latency in 1..ConnectionQuality.FAIR.maxLatencyMs -> ConnectionQuality.FAIR
            bandwidth.downloadKbps >= ConnectionQuality.POOR.minBandwidthKbps -> ConnectionQuality.POOR
            _isOnline.value -> ConnectionQuality.UNKNOWN
            else -> ConnectionQuality.OFFLINE
        }

        val previousQuality = _connectionQuality.value
        if (quality != previousQuality) {
            _connectionQuality.value = quality
            _networkEvents.tryEmit(NetworkEvent.QualityChanged(previousQuality, quality))
            logger.d(TAG, "Quality changed: $previousQuality -> $quality")
        }
    }

    @SuppressLint("NewApi")
    private fun checkDataUsage() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.M) return

        try {
            val nsm = networkStatsManager ?: return
            val endTime = System.currentTimeMillis()
            val startTime = endTime - TimeUnit.DAYS.toMillis(30)

            var wifiDl = 0L; var wifiUl = 0L
            var mobileDl = 0L; var mobileUl = 0L

            try {
                val wifiBucket = nsm.querySummaryForDevice(
                    ConnectivityManager.TYPE_WIFI, null, startTime, endTime
                )
                wifiDl = wifiBucket?.rxBytes ?: 0L
                wifiUl = wifiBucket?.txBytes ?: 0L
            } catch (_: Exception) {}

            try {
                val mobileBucket = nsm.querySummaryForDevice(
                    ConnectivityManager.TYPE_MOBILE, null, startTime, endTime
                )
                mobileDl = mobileBucket?.rxBytes ?: 0L
                mobileUl = mobileBucket?.txBytes ?: 0L
            } catch (_: Exception) {}

            _dataUsage.value = AppDataUsage(
                wifiDownloadBytes = wifiDl,
                wifiUploadBytes = wifiUl,
                mobileDownloadBytes = mobileDl,
                mobileUploadBytes = mobileUl,
                lastResetTimestamp = startTime
            )

            lastDataUsageCheckTime.set(System.currentTimeMillis())
        } catch (e: Exception) {
            logger.w(TAG, "Data usage check failed", e)
        }
    }

    private fun trackBytesTransferred() {
        try {
            val totalDl = TrafficStats.getTotalRxBytes()
            val totalUl = TrafficStats.getTotalTxBytes()

            if (totalDl != TrafficStats.UNSUPPORTED) {
                totalBytesDownloaded.set(totalDl)
            }
            if (totalUl != TrafficStats.UNSUPPORTED) {
                totalBytesUploaded.set(totalUl)
            }

            val uid = android.os.Process.myUid()
            val appDl = TrafficStats.getUidRxBytes(uid)
            val appUl = TrafficStats.getUidTxBytes(uid)

            if (appDl != TrafficStats.UNSUPPORTED) {
                sessionBytesDownloaded.set(appDl)
            }
            if (appUl != TrafficStats.UNSUPPORTED) {
                sessionBytesUploaded.set(appUl)
            }
        } catch (e: Exception) {
            logger.w(TAG, "Traffic stats tracking failed", e)
        }
    }

    private fun updateUptime() {
        uptimeSeconds.incrementAndGet()
    }

    private fun handleNetworkSwitch(from: NetworkType, to: NetworkType) {
        val event = NetworkSwitchEvent(
            timestamp = System.currentTimeMillis(),
            fromType = from,
            toType = to,
            reason = "system"
        )
        networkSwitchHistory.addLast(event)
        if (networkSwitchHistory.size > 20) networkSwitchHistory.removeFirst()

        lastNetworkSwitchTime.set(System.currentTimeMillis())
        _networkEvents.tryEmit(NetworkEvent.Switched(from, to))

        logger.i(TAG, "Network switched: $from -> $to")
    }

    private fun shouldCheckCaptivePortal(): Boolean {
        return System.currentTimeMillis() - lastCaptivePortalCheckTime.get() > 120_000L
    }

    private fun shouldEstimateBandwidth(): Boolean {
        return System.currentTimeMillis() - lastBandwidthEstimateTime.get() > 300_000L
    }

    private fun shouldRunDiagnostics(): Boolean {
        return System.currentTimeMillis() - lastDiagnosticTime.get() > 600_000L
    }

    private fun shouldCheckDataUsage(): Boolean {
        return System.currentTimeMillis() - lastDataUsageCheckTime.get() > 900_000L
    }

    private fun updateAllStates(state: ComprehensiveNetworkState) {
        _networkState.value = state
        _isOnline.value = state.isConnected && !state.isCaptivePortal
        _networkType.value = state.networkType
    }

    private fun registerCallbacks() {
        registerNetworkCallback()
        registerTelephonyCallback()
    }

    private fun unregisterCallbacks() {
        unregisterNetworkCallback()
        unregisterTelephonyCallback()
    }

    @SuppressLint("MissingPermission")
    private fun registerNetworkCallback() {
        try {
            unregisterNetworkCallback()

            val request = NetworkRequest.Builder()
                .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
                .build()

            connectivityNetworkCallback = ConnectivityNetworkCallback()

            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                connectivityManager?.registerDefaultNetworkCallback(
                    connectivityNetworkCallback!!, mainHandler
                )
            } else {
                connectivityManager?.registerNetworkCallback(
                    request, connectivityNetworkCallback!!
                )
            }
        } catch (e: Exception) {
            logger.e(TAG, "Failed to register network callback", e)
        }
    }

    private fun unregisterNetworkCallback() {
        try {
            connectivityNetworkCallback?.let {
                connectivityManager?.unregisterNetworkCallback(it)
            }
        } catch (_: Exception) {}
        connectivityNetworkCallback = null
    }

    private fun registerTelephonyCallback() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            try {
                telephonyCallback = TelephonyCallback()
                telephonyManager?.registerTelephonyCallback(
                    context.mainExecutor, telephonyCallback!!
                )
            } catch (e: Exception) {
                logger.w(TAG, "Failed to register telephony callback", e)
            }
        }
    }

    private fun unregisterTelephonyCallback() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            try {
                telephonyCallback?.let {
                    telephonyManager?.unregisterTelephonyCallback(it)
                }
            } catch (_: Exception) {}
            telephonyCallback = null
        }
    }

    private fun registerReceivers() {
        registerWifiReceiver()
        registerBatteryReceiver()
    }

    private fun unregisterReceivers() {
        unregisterWifiReceiver()
        unregisterBatteryReceiver()
    }

    private fun registerWifiReceiver() {
        try {
            val filter = IntentFilter().apply {
                addAction(WifiManager.WIFI_STATE_CHANGED_ACTION)
                addAction(WifiManager.NETWORK_STATE_CHANGED_ACTION)
                addAction(WifiManager.RSSI_CHANGED_ACTION)
            }
            wifiBroadcastReceiver = WifiBroadcastReceiver()

            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                context.registerReceiver(
                    wifiBroadcastReceiver, filter, Context.RECEIVER_NOT_EXPORTED
                )
            } else {
                @Suppress("DEPRECATION")
                context.registerReceiver(wifiBroadcastReceiver, filter)
            }
        } catch (e: Exception) {
            logger.w(TAG, "Failed to register WiFi receiver", e)
        }
    }

    private fun unregisterWifiReceiver() {
        try {
            wifiBroadcastReceiver?.let { context.unregisterReceiver(it) }
        } catch (_: Exception) {}
        wifiBroadcastReceiver = null
    }

    private fun registerBatteryReceiver() {
        try {
            val filter = IntentFilter().apply {
                addAction(Intent.ACTION_BATTERY_LOW)
                addAction(Intent.ACTION_BATTERY_OKAY)
                addAction(Intent.ACTION_POWER_CONNECTED)
                addAction(Intent.ACTION_POWER_DISCONNECTED)
            }
            batteryReceiver = BatteryBroadcastReceiver()

            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                context.registerReceiver(
                    batteryReceiver, filter, Context.RECEIVER_NOT_EXPORTED
                )
            } else {
                @Suppress("DEPRECATION")
                context.registerReceiver(batteryReceiver, filter)
            }
        } catch (e: Exception) {
            logger.w(TAG, "Failed to register battery receiver", e)
        }
    }

    private fun unregisterBatteryReceiver() {
        try {
            batteryReceiver?.let { context.unregisterReceiver(it) }
        } catch (_: Exception) {}
        batteryReceiver = null
    }

    private inner class ConnectivityNetworkCallback : ConnectivityManager.NetworkCallback() {

        override fun onAvailable(network: Network) {
            logger.d(TAG, "Network available: ${network.networkHandle}")
            val previousHandle = currentNetworkHandle.getAndSet(network.networkHandle)
            if (previousHandle != network.networkHandle && previousHandle != -1L) {
                logger.i(TAG, "Network handle changed: $previousHandle -> ${network.networkHandle}")
            }
            scope.launch {
                val state = buildComprehensiveNetworkState()
                updateAllStates(state)
                _networkEvents.tryEmit(NetworkEvent.Connected(state.networkType))
                delay(1000)
                checkCaptivePortal()
                estimateBandwidth()
                evaluateConnectionQuality()
            }
        }

        override fun onLost(network: Network) {
            logger.w(TAG, "Network lost: ${network.networkHandle}")
            currentNetworkHandle.set(-1L)
            val reason = if (activeNetworks.isEmpty()) "all_networks_lost" else "specific_network_lost"
            activeNetworks.remove(network.networkHandle)
            if (activeNetworks.isEmpty()) {
                _networkState.value = ComprehensiveNetworkState()
                _isOnline.value = false
                _networkType.value = NetworkType.UNKNOWN
                _connectionQuality.value = ConnectionQuality.OFFLINE
                _bandwidthEstimate.value = BandwidthEstimate.UNKNOWN
                _networkEvents.tryEmit(NetworkEvent.Disconnected(reason))
            }
        }

        override fun onCapabilitiesChanged(network: Network, caps: NetworkCapabilities) {
            scope.launch {
                val state = buildComprehensiveNetworkState()
                updateAllStates(state)
                evaluateConnectionQuality()
            }
        }

        override fun onLinkPropertiesChanged(network: Network, props: LinkProperties) {
            scope.launch {
                val state = buildComprehensiveNetworkState()
                updateAllStates(state)
            }
        }

        override fun onBlockedStatusChanged(network: Network, blocked: Boolean) {
            logger.w(TAG, "Network blocked: $blocked for ${network.networkHandle}")
            if (blocked) {
                _isOnline.value = false
                _connectionQuality.value = ConnectionQuality.POOR
            }
        }
    }

    @androidx.annotation.RequiresApi(Build.VERSION_CODES.S)
    private inner class TelephonyCallback : TelephonyCallback(),
        android.telephony.TelephonyCallback.SignalStrengthsListener,
        android.telephony.TelephonyCallback.ServiceStateListener,
        android.telephony.TelephonyCallback.DataConnectionStateListener {

        override fun onSignalStrengthsChanged(signalStrength: SignalStrength) {
            scope.launch {
                val state = buildComprehensiveNetworkState()
                _networkState.update { it.copy(signalStrength = state.signalStrength, signalStrengthDbm = state.signalStrengthDbm) }
            }
        }

        override fun onServiceStateChanged(serviceState: ServiceState) {
            scope.launch {
                val state = buildComprehensiveNetworkState()
                updateAllStates(state)
            }
        }

        override fun onDataConnectionStateChanged(state: Int, networkType: Int) {
            scope.launch {
                val ns = buildComprehensiveNetworkState()
                updateAllStates(ns)
                if (state == TelephonyManager.DATA_CONNECTED) {
                    _networkEvents.tryEmit(NetworkEvent.Connected(ns.networkType))
                }
            }
        }
    }

    private inner class WifiBroadcastReceiver : BroadcastReceiver() {
        override fun onReceive(context: Context?, intent: Intent?) {
            if (intent == null) return
            scope.launch {
                when (intent.action) {
                    WifiManager.WIFI_STATE_CHANGED_ACTION -> {
                        delay(2000)
                        val state = buildComprehensiveNetworkState()
                        updateAllStates(state)
                    }
                    WifiManager.NETWORK_STATE_CHANGED_ACTION -> {
                        delay(1000)
                        val state = buildComprehensiveNetworkState()
                        updateAllStates(state)
                        evaluateConnectionQuality()
                    }
                    WifiManager.RSSI_CHANGED_ACTION -> {
                        val state = buildComprehensiveNetworkState()
                        _networkState.update {
                            it.copy(wifiRssi = state.wifiRssi, signalStrength = state.signalStrength)
                        }
                    }
                }
            }
        }
    }

    private inner class BatteryBroadcastReceiver : BroadcastReceiver() {
        override fun onReceive(context: Context?, intent: Intent?) {
            if (intent == null) return
            scope.launch {
                when (intent.action) {
                    Intent.ACTION_BATTERY_LOW -> {
                        _networkState.update { it.copy(isBatterySaverActive = true) }
                        _networkEvents.tryEmit(NetworkEvent.BatterySaverToggled(true))
                    }
                    Intent.ACTION_BATTERY_OKAY -> {
                        _networkState.update { it.copy(isBatterySaverActive = false) }
                        _networkEvents.tryEmit(NetworkEvent.BatterySaverToggled(false))
                    }
                    Intent.ACTION_POWER_CONNECTED -> {
                        _networkState.update { it.copy(isBatterySaverActive = false) }
                    }
                }
            }
        }
    }

    fun recordBytesTransferred(downloaded: Long, uploaded: Long) {
        if (downloaded > 0) sessionBytesDownloaded.addAndGet(downloaded)
        if (uploaded > 0) sessionBytesUploaded.addAndGet(uploaded)
    }

    fun getNetworkDiagnosticHistory(): List<NetworkDiagnosticResult> {
        return diagnosticHistory.toList()
    }

    fun getNetworkSwitchHistory(): List<NetworkSwitchEvent> {
        return networkSwitchHistory.toList()
    }

    fun getLatencyHistory(): List<LatencySample> {
        return latencyHistory.toList()
    }

    fun shutdown() {
        logger.i(TAG, "Shutting down NetworkStateMonitor")
        stopMonitoring()
        latencyHistory.clear()
        bandwidthSamples.clear()
        packetLossHistory.clear()
        networkSwitchHistory.clear()
        diagnosticHistory.clear()
        activeNetworks.clear()
        scope.cancel()
        logger.i(TAG, "NetworkStateMonitor shutdown complete")
    }

    companion object {
        private const val TAG = "NetworkStateMonitor"
    }
}

private fun NetworkType.isMobile(): Boolean {
    return this in listOf(
        NetworkType.MOBILE, NetworkType.MOBILE_2G, NetworkType.MOBILE_3G,
        NetworkType.MOBILE_4G, NetworkType.MOBILE_5G
    )
}

private val NetworkType.displayName: String
    get() = when (this) {
        NetworkType.WIFI -> "WiFi"
        NetworkType.MOBILE -> "Mobile"
        NetworkType.MOBILE_2G -> "2G"
        NetworkType.MOBILE_3G -> "3G"
        NetworkType.MOBILE_4G -> "4G"
        NetworkType.MOBILE_5G -> "5G"
        NetworkType.ETHERNET -> "Ethernet"
        NetworkType.VPN -> "VPN"
        NetworkType.BLUETOOTH -> "Bluetooth"
        else -> "Unknown"
    }