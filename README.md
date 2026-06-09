# board — 익명 커뮤니티 게시판

> **AI Agent 팀 주도 + Human-in-the-Loop** 방식으로 개발되는 오픈소스 익명 커뮤니티입니다.

| 환경 | URL |
|------|-----|
| Production | [board.abyssworks.dev](https://board.abyssworks.dev) |
| Preview | `preview-board-{PR_NUMBER}.abyssworks.dev` |

---

## 기술 스택

| 영역 | 스택 |
|------|------|
| Backend | Go 1.25, `net/http`, `lib/pq` |
| Frontend | Vue 3 + TypeScript + Vite (Go `embed.FS` 임베디드) |
| Database | PostgreSQL 16 |
| Infra | Docker → kind (Kubernetes in Docker) → Cloudflare |

---

## 팀 구조

이 프로젝트는 **Hermes Agent** 기반의 멀티에이전트 시스템으로 운영되며, 각 에이전트가 역할을 분담합니다.

| 역할 | 에이전트 | 설명 |
|------|---------|------|
| **Commander** | Hermes Agent | 전체 오케스트레이션, 최종 검증, 사용자와의 인터페이스 |
| **PM** | Hermes Agent | 스펙 정의, 작업 분해, ADR 작성 |
| **Developer** | Hermes Agent | 코드 구현, 테스트, PR 생성 |
| **Reviewer** | Hermes Agent | 코드 리뷰, 품질 게이트, 보안 검사 |
| **DevOps** | Hermes Agent | CI/CD, Docker 빌드, k8s 배포, 인프라 관리 |

모든 결정과 결과물은 **사용자(Human-in-the-Loop)** 가 최종 승인하며, 에이전트는 제안과 실행을 담당합니다.

---

## 개발 워크플로우

```
사용자 요청 → Commander → PM(스펙) → Dev(PR) → Review → DevOps(배포) → Commander(검증) → 사용자 확인
```

### 브랜치 전략

```
main  ───── 릴리즈 브랜치 (PR merge → auto-deploy)
  └── dev ── 개발 브랜치 (feature 취합)
       └── feat/*, fix/* ── 기능 작업 브랜치 (PR Preview)
```

### 실제 배포 흐름

| 단계 | 설명 |
|------|------|
| ① 기능 작업 | `dev`에서 `feat/*` 브랜치 생성 → 개발 |
| ② PR 생성 | `feat/*` → `dev` PR → 자동 Preview 배포 |
| ③ 리뷰 | Reviewer 검토 → 승인 → `dev` 머지 |
| ④ 릴리즈 PR | `dev` → `main` PR → Review → 머지 |
| ⑤ 자동 배포 | `main` push → `board:prod` 이미지 빌드 → k8s rolling update |
| ⑥ 버전 릴리즈 | `v*` 태그 push → GitHub Release + 버전 이미지 |

---

## 시작하기

```bash
# 로컬 개발
docker run -d --name postgres -e POSTGRES_DB=community -e POSTGRES_PASSWORD=*** -p 5432:5432 postgres:16
DB_HOST=localhost go run .

# 프론트엔드 개발 (추후)
cd frontend && npm run dev

# Docker 빌드
docker build -t board:prod .
```

---

## 라이선스

MIT
