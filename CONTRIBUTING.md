# Contributing

## Setup

git clone https://github.com/wispmail/wispmail.git
cd wispmail
make init
docker compose -f build/docker-compose.yml up -d
cp .env.example .env
make dev

## Commits

Format: type(scope): description
Types: feat, fix, docs, style, refactor, perf, test, chore, ci, build, revert, security

## PR Process

1. Fork and branch
2. Write code with tests
3. make lint && make test
4. Create PR
5. Wait for review

## Code Style

Effective Go | gofmt, goimports | fmt.Errorf("context: %w", err)
Max function: 80 lines | Max complexity: 15