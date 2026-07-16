cmake_minimum_required(VERSION 3.22.1)
project("securemail_crypto")

# ============================================================================
# C++ Standard Configuration
# ============================================================================
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# ============================================================================
# Build Options
# ============================================================================
option(ENABLE_ASAN "Enable Address Sanitizer" OFF)
option(ENABLE_UBSAN "Enable Undefined Behavior Sanitizer" OFF)
option(ENABLE_TSAN "Enable Thread Sanitizer" OFF)
option(ENABLE_HARDENING "Enable Security Hardening" ON)
option(BUILD_TESTS "Build test binaries" OFF)
option(USE_OPENSSL "Use OpenSSL instead of built-in" OFF)

# ============================================================================
# Compiler Flags - Security First
# ============================================================================
set(SECURITY_COMPILE_FLAGS
    -fstack-protector-strong
    -fPIC
    -fPIE
    -D_FORTIFY_SOURCE=2
    -Wformat=2
    -Wformat-security
    -Werror=format-security
    -Werror=implicit-function-declaration
    -Werror=return-type
    -Werror=int-conversion
    -Werror=incompatible-pointer-types
    -Wno-unused-parameter
    -Wno-missing-field-initializers
)

set(OPTIMIZATION_FLAGS
    -O3
    -funroll-loops
    -fomit-frame-pointer
    -fstrict-aliasing
    -ffast-math
    -flto=thin
)

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${SECURITY_COMPILE_FLAGS}")
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} ${SECURITY_COMPILE_FLAGS}")

if(CMAKE_BUILD_TYPE STREQUAL "Release")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${OPTIMIZATION_FLAGS}")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} ${OPTIMIZATION_FLAGS}")
endif()

if(CMAKE_BUILD_TYPE STREQUAL "Debug")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -g -O0 -DDEBUG")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -g -O0 -DDEBUG")
endif()

if(ENABLE_ASAN)
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -fsanitize=address -fno-omit-frame-pointer")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fsanitize=address -fno-omit-frame-pointer")
endif()

if(ENABLE_UBSAN)
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -fsanitize=undefined")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fsanitize=undefined")
endif()

if(ENABLE_HARDENING)
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -D_FORTIFY_SOURCE=2 -fstack-clash-protection")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -D_FORTIFY_SOURCE=2 -fstack-clash-protection")
    set(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} -Wl,-z,relro -Wl,-z,now -Wl,-z,noexecstack")
    set(CMAKE_SHARED_LINKER_FLAGS "${CMAKE_SHARED_LINKER_FLAGS} -Wl,-z,relro -Wl,-z,now -Wl,-z,noexecstack")
endif()

# ============================================================================
# Platform Detection
# ============================================================================
if(ANDROID)
    message(STATUS "Building for Android: ${ANDROID_ABI}")
    set(PLATFORM_ANDROID ON)
    find_library(log-lib log)
    find_library(android-lib android)
    set(PLATFORM_LIBS ${log-lib} ${android-lib})
elseif(APPLE)
    message(STATUS "Building for Apple platform")
    set(PLATFORM_APPLE ON)
    set(PLATFORM_LIBS "")
elseif(UNIX)
    message(STATUS "Building for Linux/Unix")
    set(PLATFORM_LINUX ON)
    set(PLATFORM_LIBS pthread dl)
elseif(WIN32)
    message(STATUS "Building for Windows")
    set(PLATFORM_WINDOWS ON)
    set(PLATFORM_LIBS ws2_32 crypt32 bcrypt)
endif()

# ============================================================================
# CPU Feature Detection
# ============================================================================
include(CheckCXXCompilerFlag)

check_cxx_compiler_flag("-maes" COMPILER_SUPPORTS_AES)
check_cxx_compiler_flag("-msse4.1" COMPILER_SUPPORTS_SSE41)
check_cxx_compiler_flag("-mavx2" COMPILER_SUPPORTS_AVX2)
check_cxx_compiler_flag("-msha" COMPILER_SUPPORTS_SHA)
check_cxx_compiler_flag("-marm" COMPILER_SUPPORTS_ARM)
check_cxx_compiler_flag("-march=armv8-a+crypto" COMPILER_SUPPORTS_ARM_CRYPTO)

if(COMPILER_SUPPORTS_AES AND COMPILER_SUPPORTS_SSE41)
    set(AES_NI_FLAGS -maes -msse4.1)
    add_definitions(-DHAVE_AES_NI)
    message(STATUS "AES-NI support enabled")
endif()

if(COMPILER_SUPPORTS_AVX2)
    set(AVX2_FLAGS -mavx2)
    add_definitions(-DHAVE_AVX2)
    message(STATUS "AVX2 support enabled")
endif()

if(COMPILER_SUPPORTS_SHA)
    set(SHA_NI_FLAGS -msha)
    add_definitions(-DHAVE_SHA_NI)
    message(STATUS "SHA-NI support enabled")
endif()

if(COMPILER_SUPPORTS_ARM_CRYPTO AND ANDROID)
    set(ARM_CRYPTO_FLAGS -march=armv8-a+crypto)
    add_definitions(-DHAVE_ARM_CRYPTO)
    message(STATUS "ARM Crypto Extensions enabled")
endif()

# ============================================================================
# External Dependencies
# ============================================================================
if(USE_OPENSSL)
    find_package(OpenSSL REQUIRED)
    message(STATUS "Using OpenSSL: ${OPENSSL_VERSION}")
    add_definitions(-DUSE_OPENSSL)
endif()

# ============================================================================
# Source Files Organization
# ============================================================================
set(CRYPTO_SOURCES
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/aes_wrapper.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/argon2_wrapper.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/base64_wrapper.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/hash_wrapper.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/key_derivation.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/random_wrapper.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/sha_wrapper.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/chacha20_wrapper.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/poly1305_wrapper.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/curve25519_wrapper.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/ed25519_wrapper.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/hmac_wrapper.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/pbkdf2_wrapper.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto/blake2b_wrapper.cpp
)

set(JNI_SOURCES
    ${CMAKE_CURRENT_SOURCE_DIR}/jni/jni_bridge.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/jni/jni_crypto.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/jni/jni_utils.cpp
)

set(UTILS_SOURCES
    ${CMAKE_CURRENT_SOURCE_DIR}/utils/buffer_utils.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/utils/string_utils.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/utils/memory_utils.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/utils/timing_utils.cpp
)

set(ALL_SOURCES
    ${CRYPTO_SOURCES}
    ${JNI_SOURCES}
    ${UTILS_SOURCES}
)

# ============================================================================
# Library Target
# ============================================================================
add_library(securemail_crypto SHARED ${ALL_SOURCES})

target_include_directories(securemail_crypto
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}
        ${CMAKE_CURRENT_SOURCE_DIR}/crypto
        ${CMAKE_CURRENT_SOURCE_DIR}/jni
        ${CMAKE_CURRENT_SOURCE_DIR}/utils
)

target_link_libraries(securemail_crypto
    ${PLATFORM_LIBS}
)

if(USE_OPENSSL)
    target_link_libraries(securemail_crypto OpenSSL::SSL OpenSSL::Crypto)
    target_compile_definitions(securemail_crypto PRIVATE USE_OPENSSL)
endif()

# Apply CPU-specific flags
if(AES_NI_FLAGS AND NOT ARM)
    set_source_files_properties(
        ${CMAKE_CURRENT_SOURCE_DIR}/crypto/aes_wrapper.cpp
        PROPERTIES COMPILE_FLAGS "${AES_NI_FLAGS}"
    )
endif()

if(AVX2_FLAGS AND NOT ARM)
    set_source_files_properties(
        ${CMAKE_CURRENT_SOURCE_DIR}/crypto/hash_wrapper.cpp
        ${CMAKE_CURRENT_SOURCE_DIR}/crypto/sha_wrapper.cpp
        PROPERTIES COMPILE_FLAGS "${AVX2_FLAGS}"
    )
endif()

if(ARM_CRYPTO_FLAGS AND ANDROID)
    set_source_files_properties(
        ${CMAKE_CURRENT_SOURCE_DIR}/crypto/aes_wrapper.cpp
        ${CMAKE_CURRENT_SOURCE_DIR}/crypto/sha_wrapper.cpp
        PROPERTIES COMPILE_FLAGS "${ARM_CRYPTO_FLAGS}"
    )
endif()

# ============================================================================
# Target Properties
# ============================================================================
set_target_properties(securemail_crypto PROPERTIES
    CXX_STANDARD 20
    CXX_STANDARD_REQUIRED ON
    CXX_EXTENSIONS OFF
    POSITION_INDEPENDENT_CODE ON
    VISIBILITY_INLINES_HIDDEN ON
    CXX_VISIBILITY_PRESET hidden
    LINK_DEPENDS_NO_SHARED ON
)

if(ANDROID)
    set_target_properties(securemail_crypto PROPERTIES
        ANDROID_SKIP_ANTI_VIRUS true
    )
endif()

# ============================================================================
# Install Rules
# ============================================================================
install(TARGETS securemail_crypto
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    RUNTIME DESTINATION bin
)

install(DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/
    DESTINATION include/securemail_crypto
    FILES_MATCHING PATTERN "*.h"
    PATTERN "CMakeLists.txt" EXCLUDE
)

# ============================================================================
# Test Target
# ============================================================================
if(BUILD_TESTS)
    enable_testing()
    
    set(TEST_SOURCES
        ${CMAKE_CURRENT_SOURCE_DIR}/../test/cpp/test_main.cpp
        ${CMAKE_CURRENT_SOURCE_DIR}/../test/cpp/test_aes.cpp
        ${CMAKE_CURRENT_SOURCE_DIR}/../test/cpp/test_hash.cpp
        ${CMAKE_CURRENT_SOURCE_DIR}/../test/cpp/test_argon2.cpp
        ${CMAKE_CURRENT_SOURCE_DIR}/../test/cpp/test_random.cpp
        ${CMAKE_CURRENT_SOURCE_DIR}/../test/cpp/test_curve25519.cpp
        ${CMAKE_CURRENT_SOURCE_DIR}/../test/cpp/test_ed25519.cpp
    )
    
    add_executable(securemail_crypto_test ${TEST_SOURCES})
    
    target_link_libraries(securemail_crypto_test
        securemail_crypto
        ${PLATFORM_LIBS}
    )
    
    target_include_directories(securemail_crypto_test
        PRIVATE
            ${CMAKE_CURRENT_SOURCE_DIR}
            ${CMAKE_CURRENT_SOURCE_DIR}/../test/cpp
    )
    
    add_test(NAME securemail_crypto_test COMMAND securemail_crypto_test)
endif()

# ============================================================================
# Print Configuration Summary
# ============================================================================
message(STATUS "========================================")
message(STATUS "SecureMail Crypto Library Configuration")
message(STATUS "========================================")
message(STATUS "Build Type:       ${CMAKE_BUILD_TYPE}")
message(STATUS "C++ Standard:     ${CMAKE_CXX_STANDARD}")
message(STATUS "Platform:         ${CMAKE_SYSTEM_NAME}")
message(STATUS "AES-NI:           ${COMPILER_SUPPORTS_AES}")
message(STATUS "AVX2:             ${COMPILER_SUPPORTS_AVX2}")
message(STATUS "SHA-NI:           ${COMPILER_SUPPORTS_SHA}")
message(STATUS "ARM Crypto:       ${COMPILER_SUPPORTS_ARM_CRYPTO}")
message(STATUS "ASAN:             ${ENABLE_ASAN}")
message(STATUS "Hardening:        ${ENABLE_HARDENING}")
message(STATUS "OpenSSL:          ${USE_OPENSSL}")
message(STATUS "Tests:            ${BUILD_TESTS}")
message(STATUS "========================================")