# AGENTS.md — board 프로젝트

> 이 문서는 board 프로젝트의 기술 스택, 구조, 컨벤션을 정의한다.
> SOUL.md에서 정의한 역할별 원칙 위에 이 문서의 내용이 프로젝트 레벨에서 적용된다.

---

## 프로젝트 개요

| 항목 | 내용 |
|------|------|
| 저장소 | `abyss-works/board` |
| 언어 | Go 1.22 |
| DB | PostgreSQL (lib/pq) |
| API | REST (net/http, no framework) |
| 프론트엔드 | 임베디드 정적 파일 (`embed.FS`) |
| 빌드 | `go build -o board .` |
| 배포 | Docker → kind k8s 클러스터 |

## 디렉토리 구조

```
board/
├── main.go              # 진입점, 라우팅, 핸들러
├── frontend/dist/        # 임베디드 프론트엔드 빌드 결과물
├── Dockerfile            # 멀티스테이지 빌드
├── k8s/ 또는 ./
│   ├── app.yaml          # Deployment + Service
│   ├── postgres.yaml     # PostgreSQL StatefulSet
│   └── ingress.yaml      # Ingress
├── docs/spec/            # PM 작성 스펙 문서
├── docs/adr/             # Architecture Decision Records
├── AGENTS.md             # (이 파일)
└── CLAUDE.md
```

## 빌드 & 테스트

```bash
go build -o board .        # 빌드
go vet ./...                          # 정적 분석
# 테스트: (현재 테스트 파일 없음 — 추후 추가)
```

## API 패턴

- 표준 `net/http` 핸들러
- JSON 요청/응답
- 라우팅: `http.HandleFunc` with path prefix matching
- DB: `database/sql` + `lib/pq`
- 포트: 환경변수 `PORT` (기본 8080)
- DB 연결: `DATABASE_URL` 환경변수

## 배포 (k8s)

- kind 클러스터 (`abyssworks`)
- 네임스페이스: `default`
- Ingress: `*.abyssworks.dev` (Cloudflare Proxy → kind ingress-nginx)
- PostgreSQL: 동일 클러스터 내 StatefulSet

## 커밋 컨벤션

- 타입: `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `test:`
- 제목: 한글 30자 이내, 명사형 종결
- 본문: 불릿 리스트, 파일 단위 변경 추적
- 커밋 단위: 원자적, 계층별 분리 (Data → API → UI → Docs)
