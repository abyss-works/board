# Infrastructure

> 인프라, 배포, Git 워크플로우.

---

# 프로젝트 실행

> 프로젝트별 실제 실행 방법.

## 로컬 개발

1. **DB 실행**: `docker run -d --name postgres -e POSTGRES_DB=community -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:16`
2. **백엔드 실행**: `DB_HOST=localhost go run .`
3. **프론트엔드 실행**: `cd frontend && npm run dev` (추후 도입)

## Kubernetes 배포

```bash
kubectl apply -f postgres.yaml
kubectl apply -f app.yaml
kubectl apply -f ingress.yaml
```

---

# Git 브랜치 전략

| 구분 | 설명 |
|------|------|
| 메인 브랜치 | `main` — 프로젝트 기본 브랜치 |
| 작업 브랜치 | `task/{번호}` (예: `task/1`, `task/2`) |
| 커밋 컨벤션 | [commit-convention.md](../commit-convention.md) 참조 |
| 커밋 형식 | `[TASK-{번호}] {type}: {한글 설명}` |
| 커밋 타입 | `init`, `feat`, `fix`, `refactor`, `chore`, `docs`, `style`, `test`, `clean` |

## 브랜치 생성 및 병합 규칙

- 작업 시작 전 반드시 메인 브랜치에서 새 작업 브랜치를 생성한다.
- 작업 완료 후 PR(Pull Request)을 통해 메인 브랜치에 병합한다.
- 직접 메인 브랜치에 push하는 행위는 절대 금지한다.

---

# 환경 변수 파일

민감 정보가 포함될 수 있는 환경 변수 파일은 절대 버전 관리에 포함해서는 안 된다:

| 파일 | 용도 | Git 추적 |
|------|------|----------|
| `.env` | 공개 환경 변수 (DB_HOST, DB_PORT 등) | **금지** (`.gitignore` 등록) |
| `.secret.env` | 비밀 환경 변수 (DB_PASSWORD 등) | **절대 금지** |
| `.env.example` | `.env` 템플릿 | Git 추적 |
| `.secret.env.example` | `.secret.env` 템플릿 (실제 값 제외) | Git 추적 |

## 현재 환경 변수 목록

| 변수 | 기본값 | 설명 |
|------|------|------|
| `DB_HOST` | `postgres` | PostgreSQL 호스트 |
| `DB_PORT` | `5432` | PostgreSQL 포트 |
| `DB_USER` | `postgres` | DB 사용자 |
| `DB_PASSWORD` | `postgres` | DB 비밀번호 |
| `DB_NAME` | `community` | DB 이름 |
| `PORT` | `8080` | 서버 포트 |

---

# 컨테이너 오케스트레이션

## 서비스 구성

| 파일 | 내용 |
|------|------|
| `postgres.yaml` | PostgreSQL StatefulSet + Service |
| `app.yaml` | community-board Deployment + Service |
| `ingress.yaml` | Ingress 라우팅 |
| `Dockerfile` | 멀티스테이지 Go 빌드 → 경량 실행 이미지 |

## Dockerfile 구성

- 빌드 스테이지: `golang:1.22-alpine` → `go build`
- 실행 스테이지: `alpine:3.19`
- `EXPOSE 8080`
- 비특권 사용자로 실행

---

# 빌드

- **백엔드**: `go build -o community-board .` → 단일 바이너리 생성
- **프론트엔드**: `npm run build` → `dist/` 정적 파일 생성 (추후)
- **Docker**: `docker build -t community-board .`

---

# CI/CD

CI/CD 파이프라인 구성 (추후 도입):

1. **테스트**: `go test ./...`
2. **린트**: `golangci-lint run`
3. **빌드**: Docker 이미지 빌드 및 레지스트리 푸시
4. **배포**: `kubectl apply`

---

# 보안 정책

- `.env`, `.secret.env` 파일은 어떠한 경우에도 버전 관리에 포함해서는 안 된다. 이 규칙을 위반할 경우 보안 사고로 간주한다.
- CI/CD 시크릿은 환경 변수 파일이 아닌 CI/CD 도구의 시크릿 관리 기능을 사용한다.
- Docker 볼륨 데이터(`db_data`, `logs` 등)는 `.gitignore`에 추가하여 실수로 커밋되지 않도록 한다.

---

# 주의사항

- Docker 볼륨 데이터는 `.gitignore`에 추가하여 실수로 커밋되지 않도록 한다.
- 실행 스크립트는 OS별로 제공하고, 실행 전 환경 변수 파일 존재 여부를 검증하는 로직을 포함한다.
- CI/CD 시크릿은 환경 변수 파일이 아닌 CI/CD 도구의 시크릿 관리 기능을 사용한다.
