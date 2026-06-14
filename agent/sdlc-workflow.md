# SDLC Workflow — 소프트웨어 개발 생애주기 워크플로우

> 이 문서는 abyss-works 프로젝트의 전체 SDLC(Software Development Life Cycle) 워크플로우를 정의한다.
> 모든 에이전트와 개발자는 이 워크플로우를 준수해야 한다.

---

## 1. 워크플로우 개요

### 코어 원칙

| 원칙 | 설명 |
|------|------|
| **Phase = PR** | 하나의 논리적 기능/개선 단위는 하나의 Pull Request로 대응된다 |
| **Atomic unit = Commit** | 하나의 원자적 변경은 하나의 커밋으로 대응된다 |
| **의존성 역순 커밋** | 데이터 계층 → 비즈니스 로직 → API → 클라이언트 → UI 순서로 커밋 |
| **문서 선행** | 코드 변경 전에 스펙 문서(`docs/spec/`) 또는 계획을 먼저 수립 |
| **Human-in-the-Loop** | 모든 PR의 최종 머지 결정권은 사용자에게 있음 |
| **변경 확인 후 실행** | 파일/함수/경로는 `grep`으로 실제 존재 확인 후 조작. 추측 금지. |

### 전체 흐름

```
사용자 요청 또는 Issue
  │
  ▼
① 현황 분석 (코드베이스 + Git + GitHub 상태 종합 파악)
  │
  ▼
② Phase 계획 수립 (작은 단위로 분할, 우선순위 지정)
  │
  ▼
③ Phase 실행 (반복):
  ├── a. 작업 브랜치 생성 (dev base, feat/{설명})
  ├── b. 원자적 커밋 N개 (의존성 역순, 컨벤셔널 커밋)
  ├── c. dev push → PR 생성 (dev → main)
  ├── d. PR 바디 작성 (변경사항 섹션별 정리, Issue 연결)
  └── e. 사용자에게 보고 (PR 링크 + 요약)
  │
  ▼
④ 문서화 (배운 점, 변경된 워크플로우 기록, 필요시 CLAUDE.md 업데이트)
```

---

## 2. Phase 계획

### Phase란?

Phase는 **하나의 독립적인 PR로 배포 가능한 작업 단위**다.
하나의 Phase가 머지되지 않아도 다음 Phase의 설계/코딩에 지장이 없어야 한다.

### Phase 분할 기준

| 기준 | 설명 |
|------|------|
| **독립성** | Phase 간 의존성을 최소화. 각 Phase는 이전 Phase와 무관하게 리뷰 가능 |
| **크기** | 1~6개 커밋. 1인-1일 이내에 완료 가능한 규모 |
| **일관성** | 하나의 Phase는 하나의 주제(기능/버그/UI/인프라)만 다룸 |
| **역할 혼합 금지** | 하나의 Phase에 BE + FE + DevOps 변경을 섞지 않음 (단, BE+FE는 같은 주제면 허용) |
| **우선순위** | 버그 > UX 개선 > 기능 개발 > 코드 품질 > 인프라 |

### Phase 명명 규칙

```
Phase N: {행위} {대상}
```

예시:
- `Phase 0: feat/comment-edit-delete 머지`
- `Phase 1: 비밀번호 UX 개선`
- `Phase 2: 게시글 목록 UX 개선`

---

## 3. 브랜치 전략

### 브랜치 계층

```
main  ────────────────────── 릴리즈 브랜치
  │                          PR merge → 자동배포
  │
  └── dev ────────────────── 개발 취합 브랜치
       │                     feature 브랜치의 목적지
       │
       ├── feat/*            기능 개발 브랜치
       ├── fix/*             버그 수정 브랜치
       ├── chore/*           설정/의존성 브랜치
       ├── refactor/*        리팩토링 브랜치
       ├── docs/*            문서 작업 브랜치
       └── test/*            검증/테스트 브랜치
```

### 브랜치 운영 규칙

| 규칙 | 설명 |
|------|------|
| **모든 작업 브랜치는 `dev`에서 분기** | `main`에서 직접 분기 금지 |
| **작업 브랜치명** | `{type}/{간략-설명}` (케밥 케이스, 영문) |
| **작업 브랜치 수명** | PR 머지 즉시 삭제 |
| **dev → main PR** | 릴리즈 단위. 다수의 feature를 취합하여 배포 |
| **main 직접 push 금지** | 모든 변경은 PR을 통해서만 |

### 예외: 기존 브랜치가 main 기반일 경우

실제 사례: `feat/comment-edit-delete` 브랜치가 `main`에서 분기되어 있었다.

**처리 방법:**
```bash
# 1. merge-base 확인
git merge-base main feat/comment-edit-delete
git merge-base dev feat/comment-edit-delete

# 2. 양쪽의 커밋 차이 확인
git log --oneline feat/comment-edit-delete..dev   # dev만 가진 커밋
git log --oneline dev..feat/comment-edit-delete    # feat만 가진 커밋

# 3. 충돌 가능성 낮으면 --no-ff merge로 dev에 병합
git checkout dev
git merge feat/comment-edit-delete --no-ff -m "chore: {브랜치명} → dev merge"
```

### 머지 흐름

```
feat/xyz ──PR──→ dev ──PR──→ main ──v* 태그──→ Release
    (Preview)      (취합)       (배포)         (GitHub Release)
```

---

## 4. 커밋 워크플로우 (실제 적용 상세)

### 4.1 원자적 커밋

**하나의 커밋 = 하나의 논리적 변경.** 다음 기준을 충족해야 한다:

| 기준 | 설명 |
|------|------|
| **단일 책임** | 하나의 커밋은 하나의 계층 또는 하나의 관심사만 변경 |
| **독립성** | 각 커밋은 이론적으로 단독으로 리뷰 가능 |
| **일관성** | 빌드가 깨지지 않는 상태여야 함 |

### 4.2 커밋 순서 (의존성 역순)

변경사항은 **의존성의 역순**으로 커밋한다. 즉, 다른 것들이 의존하는 기초부터 먼저 만든다:

```
① 데이터 계층 (Model, Schema, Migration, DB 쿼리)
  → ② 비즈니스 로직 계층 (Service, Validation, 비밀번호 검증)
    → ③ API/Controller 계층 (Handler, Router, 엔드포인트)
      → ④ 클라이언트 API 연동 (Types 정의, API 함수)
        → ⑤ 상태 관리 계층 (Store, Service)
          → ⑥ 화면/UI (View, Component, 스타일)
```

**실제 사례** (feat/comment-edit-delete의 1개 커밋):

커밋 `bd6a9dc` — 하나의 커밋이지만 내부적으로 다음 순서로 변경:
```
main.go              ← 데이터: comments 테이블 마이그레이션 (BE-1)
main.go              ← 모델: Comment.UpdatedAt 필드 추가 (BE-2)
main.go              ← API: 댓글 작성 password 확장 (BE-3)
main.go              ← API: 댓글 조회 soft-delete 필터 (BE-4)
main.go              ← API: handleCommentByID PUT/DELETE (BE-5, BE-6)
types/index.ts       ← 타입: Comment.updatedAt, UpdateCommentRequest (FE-1)
api/comments.ts      ← API: updateComment, deleteComment (FE-2)
CommentItem.vue      ← UI: 수정/삭제 버튼 + 인라인 폼 (FE-3)
PostDetailView.vue   ← UI: 비밀번호 필드 + 이벤트 핸들링 (FE-4)
```

### 4.3 커밋 메시지 상세

**형식:**
```
{type}: {한글 설명}
```

또는 머지 커밋:
```
chore: {source-branch} → {target-branch} merge

- {type}: {설명} ({sha})
- {type}: {설명} ({sha})
```

**실제 머지 커밋 사례:**
```
chore: feat/comment-edit-delete → dev merge

- feat: 댓글 수정/삭제 기능 구현 (bd6a9dc)
- fix: 러너 라벨 self-hosted로 통일 (b1ff2f7)
- chore: 워크플로우 재트리거 (1f0adf4)
```

**금지 패턴:**
- ❌ 영어 제목만 (`setup core dependencies`)
- ❌ 타입 누락 (`CI 스크립트 작성`)
- ❌ 제목에 마침표
- ❌ 본문을 한 문단으로만 작성 (불릿 리스트 필수)

---

## 5. PR 워크플로우 (상세 + 실제 사례)

### 5.1 PR 생성 명령어

```bash
# dev → main PR
gh pr create \
  --repo abyss-works/board \
  --base main \
  --head dev \
  --title "v0.2.0: 댓글 수정/삭제 기능 + 러너 라벨 통일" \
  --body "## 변경사항

### Backend
- 댓글 수정/삭제 API 구현 (PUT /api/comments/{id}, DELETE /api/comments/{id})
- 댓글 작성 시 비밀번호(bcrypt) 선택 저장
- 댓글 목록 soft-delete 필터 (WHERE deleted_at IS NULL)

### Frontend
- CommentItem 컴포넌트: 수정/삭제 버튼 + 인라인 수정 폼 + 삭제 확인 다이얼로그
- PostDetailView: 댓글 작성 시 비밀번호 입력 필드 추가

### DevOps
- 러너 라벨 [self-hosted, prod/preview] → [self-hosted] 통일

## 관련 이슈
- Closes #10

## 커밋
- merge commit: 66be6ca"
```

### 5.2 PR 제목 규칙

```
{type}: {Phase 설명} ({관련 이슈 있으면 #N})
```

예시:
- `feat: 댓글 수정/삭제 기능 구현 (#10)`
- `fix: 비밀번호 선택 시 영구 삭제 불가 문제 해결 (#8)`
- `feat: 게시글 목록 페이지네이션 추가`

### 5.3 PR 바디 상세 템플릿

```markdown
## 변경사항

### Backend
- {변경 1} (API 경로나 파일명 포함)
- {변경 2}

### Frontend
- {변경 1}
- {변경 2}

### DevOps
- {변경 1}

## 관련 이슈
- Closes #{N}         ← PR 머지 시 자동으로 Issue 닫힘
- Related to #{N}     ← 참조만 할 경우

## 커밋
- merge commit: {sha}
```

### 5.4 Issue 참조 방식 (매우 중요)

PR 본문에서 Issue를 참조할 때는 **GitHub 키워드**를 정확히 사용해야 한다.
`Closes #N` 키워드는 PR이 머지될 때 해당 Issue를 **자동으로 닫아준다**.

**키워드 종류:**

| 키워드 | 효과 | 사용처 |
|--------|------|--------|
| `Closes #N` | 머지 시 Issue 자동 닫힘 | 기능 구현, 버그 수정 |
| `Fixes #N` | 머지 시 Issue 자동 닫힘 | 버그 수정 (Closes와 동일) |
| `Resolves #N` | 머지 시 Issue 자동 닫힘 | 동일 |
| `Related to #N` | Issue를 닫지 않음 | 참조만 할 때 |
| `Ref #N` | Issue를 닫지 않음 | 참조만 할 때 |

**주의:** 본문 어디에 위치하든 GitHub이 자동 인식한다.
단, **코드 블록(\`\`\`) 안에 넣으면 인식되지 않는다.**

**올바른 예:**
```markdown
## 관련 이슈
- Closes #10

이 PR이 머지되면 댓글 수정/삭제 기능 Issue #10이 자동으로 닫힙니다.
```

### 5.5 머지 방식

| 브랜치 | 머지 방식 | 명령어 | 설명 |
|--------|-----------|--------|------|
| feat/* → dev | `--no-ff` | `git merge --no-ff feat/xyz -m "chore: ..."` | 머지 커밋 생성, 히스토리 보존 |
| dev → main | `--no-ff` | (GitHub PR 머지 버튼) | 머지 커밋으로 저장 |

**`--no-ff`를 사용하는 이유:**
- fast-forward 금지 → feature 브랜치의 존재를 히스토리에 명시적으로 남김
- 머지 커밋에 요약 메시지를 포함하여 **무슨 내용이 머지됐는지 한눈에 파악** 가능
- `git log --graph --oneline`에서 브랜치 구조를 시각적으로 확인 가능

**실제 결과 (git log):**
```
*   66be6ca (HEAD -> dev) chore: feat/comment-edit-delete → dev merge
|\
| * 1f0adf4 chore: 워크플로우 재트리거
| * b1ff2f7 fix: 러너 라벨 self-hosted로 통일
| * bd6a9dc feat: 댓글 수정/삭제 기능 구현
|/
* 93f80dd docs: 댓글 수정/삭제 기능 스펙 문서
*   82ac1bb (tag: v0.1.1, main) Merge dev → main: v0.1.1
```

### 5.6 머지 조건

| 조건 | 설명 |
|------|------|
| 사용자 승인 | Human-in-the-Loop: 사용자가 직접 머지 결정 |
| CI 통과 | GitHub Actions 워크플로우 성공 |
| 컨플릭트 없음 | 대상 브랜치와 충돌 없음. 충돌 시 `git merge {target}`으로 로컬 해결 후 push |

---

## 6. 작업 실행 워크플로우 (실제 명령어 + 출력 포함)

### 6.1 작업 시작 전 — 현황 분석

**Step 1: 저장소 브랜치 히스토리 파악**
```bash
cd /opt/data/workspaces/abyss-works/board

# 전체 브랜치 + 태그 시각화
git log --all --oneline --graph --decorate

# 브랜치 목록
git branch -a

# 각 브랜치의 최근 커밋
git log --oneline <branch> -5
```

**Step 2: GitHub 상태 파악 (gh CLI)**
```bash
# 모든 워크플로우 목록
gh workflow list --repo abyss-works/board --json id,name,path,state

# 워크플로우별 상세 (생성일, 업데이트일)
gh api repos/abyss-works/board/actions/workflows \
  --jq '.workflows[] | {id, name, state, path, created_at, updated_at}'

# 모든 PR (state: open, closed, merged, all)
gh pr list --repo abyss-works/board --state all --limit 30 \
  --json number,title,state,headRefName,baseRefName,createdAt,mergedAt,labels

# 모든 Issue
gh issue list --repo abyss-works/board --state all --limit 20 \
  --json number,title,state,labels,createdAt,body

# 특정 Issue 상세
gh issue view {N} --repo abyss-works/board --json body,labels,comments

# CI/CD 실행 현황
gh run list --repo abyss-works/board --limit 20 \
  --json databaseId,displayTitle,headBranch,status,conclusion,createdAt

# 특정 Run 상세 (Job 단계별 시간)
gh run view {run-id} --repo abyss-works/board --json jobs

# 실패한 Run 로그
gh run view {run-id} --repo abyss-works/board --log-failed

# 러너 상태
gh api repos/abyss-works/board/actions/runners \
  --jq '.runners[] | {name, os, status, busy, labels: [.labels[].name]}'
```

**Step 3: 설정 파일 읽기 (필수)**
```bash
# 최상위 헌법 문서
cat AGENTS.md
cat CLAUDE.md

# 작업 도메인별 세부 명세
cat agent/project/backend.md
cat agent/project/frontend.md
cat agent/project/infra.md

# 서브 컨벤션
cat agent/commit-convention.md
cat agent/history-logging.md
```

### 6.2 Phase 실행 — 브랜치 작업

**Step 1: dev 베이스 작업 브랜치 생성**
```bash
git checkout dev
git pull origin dev
git checkout -b feat/{phase-description}
```

**Step 2: 작업 + 원자적 커밋**

변경 전 항상 확인:
```bash
# 파일/함수/패턴이 실제로 존재하는지 grep으로 확인
grep -rn "target_function" --include="*.go"

# 변경 전 파일 구조
find . -maxdepth 3 -not -path './.git/*' -not -path '*/node_modules/*' | sort
```

커밋 (의존성 역순):
```bash
# 1차: 데이터/모델 계층
git add main.go  # 예: DB 마이그레이션, 구조체 변경
git commit -m "feat: ..."

# 2차: API 계층
git add main.go  # 예: 새 핸들러, 라우트
git commit -m "feat: ..."

# 3차: 프론트 타입
git add frontend/src/types/index.ts
git commit -m "feat: ..."

# 4차: 프론트 API
git add frontend/src/api/comments.ts
git commit -m "feat: ..."

# 5차: 프론트 UI
git add frontend/src/components/CommentItem.vue
git commit -m "feat: ..."
```

**Step 3: 머지 커밋 생성 (dev에 통합)**

기존 브랜치를 dev에 통합할 때:
```bash
# 양방향 커밋 차이 확인
git log --oneline main..dev
git log --oneline dev..feat/comment-edit-delete

# 컨플릭트 없으면 --no-ff 머지
git checkout dev
git merge feat/comment-edit-delete --no-ff \
  -m "chore: feat/comment-edit-delete → dev merge

- feat: 댓글 수정/삭제 기능 구현 (bd6a9dc)
- fix: 러너 라벨 self-hosted로 통일 (b1ff2f7)
- chore: 워크플로우 재트리거 (1f0adf4)"

# 컨플릭트 발생하면 수동 해결:
# git mergetool  또는  vim에서 직접 수정
# git add <resolved-file>
# git merge --continue

# push
git push origin dev
```

**Step 4: PR 생성**
```bash
gh pr create \
  --repo abyss-works/board \
  --base main \
  --head dev \
  --title "v0.2.0: 댓글 수정/삭제 기능 + 러너 라벨 통일" \
  --body "$(cat <<EOF
## 변경사항

### Backend
- 댓글 수정/삭제 API 구현 (...)

### Frontend
- CommentItem 컴포넌트: 수정/삭제 UI 추가

## 관련 이슈
- Closes #10

## 커밋
- merge commit: 66be6ca
EOF
)"
```

**Step 5: 사용자에게 보고**
```
## ✅ Phase N 완료

**PR #{번호}** — {설명}
🔗 {PR 링크}

| 항목 | 상태 |
|------|------|
| {작업} | ✅ |
| PR 생성 | ✅ |

리뷰/머지는 직접 해주시면 됩니다.
```

### 6.3 작업 후 — 문서화

새로운 워크플로우를 발견하거나 기존 문서가 부족하면 즉시 보강:

```bash
# 1. 새 문서 생성
touch agent/{topic}.md

# 2. 내용 작성 (마크다운, 헤더 구조, 예시 포함)

# 3. CLAUDE.md 문서 체계 테이블에 참조 추가
# CLAUDE.md의 "문서 체계" 섹션 수정
```

---

## 7. PR 본문 Issue 참조 — GitHub 마크다운 렌더링 상세

### 7.1 동작 원리

GitHub는 PR 본문에서 `Closes #N` 패턴을 감지하면:
1. PR 본문에서 해당 텍스트를 **링크로 자동 변환** (Issue 제목 + #번호)
2. Issue 페이지에 "이 PR이 머지되면 이 Issue가 닫힙니다" 배너 표시
3. PR 머지 시 Issue 자동 닫힘 + 두 항목이 서로 링크됨

### 7.2 실제 렌더링 결과

본문에 아래 내용을 포함하면:
```markdown
## 관련 이슈
- Closes #10
```

GitHub에서 다음과 같이 렌더링된다:
```
- Closes #10
  ↓ (자동 변환)
- <키워드> <링크:댓글 수정/삭제 기능 #10>
```

키워드(`Closes`)에는 초록색 `✓` 아이콘이, Issue 링크에는 `#N`과 제목이 표시된다.

### 7.3 여러 Issue 연결

```markdown
## 관련 이슈
- Closes #8
- Closes #10
- Related to #3
```

### 7.4 주의사항

- ❌ 코드 블록(```) 안에 넣으면 GitHub이 인식하지 않음
- ❌ `Close #10` (오타) → 인식 안 됨. 정확히 `Closes #10`
- ❌ `closes #10` (소문자여도 인식되나, `Closes`로 통일 권장)
- ✅ PR 본문 아무 위치나 상관없으나, **"관련 이슈" 섹션에 모아서 작성** 권장
- ✅ 머지 커밋 메시지에 넣어도 동작하지만, PR 본문에 넣는 것이 가독성 좋음

---

## 8. 머지 전략 상세

### 8.1 `--no-ff` (No Fast-Forward)

**사용 명령어:**
```bash
git merge <branch> --no-ff -m "{type}: {요약 메시지}"
```

**효과:**
- 대상 브랜치가 현재 브랜치보다 앞서 있어도 **항상 머지 커밋을 생성**
- 머지 커밋에 포함된 feature 브랜치의 존재가 히스토리에 영구 기록됨
- `git log --graph`에서 브랜치 구조를 시각화 가능

**실제 사례 (Phase 0):**
```
*   66be6ca (dev) chore: feat/comment-edit-delete → dev merge       ← 머지 커밋
|\
| * 1f0adf4 chore: 워크플로우 재트리거                              ← feature 브랜치
| * b1ff2f7 fix: 러너 라벨 self-hosted로 통일
| * bd6a9dc feat: 댓글 수정/삭제 기능 구현
|/
* 93f80dd docs: 댓글 수정/삭제 기능 스펙 문서                        ← dev의 기존 커밋
```

### 8.2 머지 vs 리베이스 vs 스쿼시

| 방식 | 특징 | 사용처 |
|------|------|--------|
| **`--no-ff` 머지** | 머지 커밋 생성, 브랜치 구조 보존 | **기본값**. 모든 머지에 사용 |
| **리베이스** | 커밋 히스토리를 선형으로 정리 | 브랜치 최신화용 (PR 전), feature 내부 정리용 |
| **스쿼시 머지** | feature의 모든 커밋을 1개로 압축 | GitHub PR 머지 옵션. **이 프로젝트에서는 사용하지 않음** |

> **이 프로젝트는 `--no-ff`를 원칙으로 한다.** 스쿼시는 feature 브랜치의 세부 작업 내역을
> 히스토리에서 지워버리므로, 문제 발생 시 원인 추적이 어려워진다.

---

## 9. 문서 우선순위 (Precedence)

충돌 시 상위 항목이 하위 항목을 오버라이드:

| 순위 | 계층 | 파일 | 설명 |
|------|------|------|------|
| **1** | 프로젝트 헌법 | `CLAUDE.md`, `AGENTS.md` | 최상위 가드레일. 위반 불가. |
| **2** | SDLC 워크플로우 | `agent/sdlc-workflow.md` | **(이 파일)** 개발 프로세스 전반 |
| **3** | 사용자 로컬 설정 | `.claude/`, `.codex/` | 사용자별 전역 설정 |
| **4** | 세션 컨텍스트 | 사용자 입력 | 대화 중 추가 지침 |
| **5** | 프로젝트 세부 명세 | `agent/project/*` | 기술스택별 컨벤션 |
| **6** | 서브 컨벤션 | `agent/commit-convention.md`, `agent/history-logging.md` | 세부 규칙 |
| **7** | 에이전트 판단 | — | 상위 지침에 없는 사항 |

---

## 10. Phase 현황 관리

모든 Phase는 `todo` 도구로 추적한다:

```markdown
Phase 0: feat/comment-edit-delete 머지
├── dev에 feat/comment-edit-delete 병합  ✅
├── dev push                               ✅
├── PR #12 생성                             ✅
└── 머지                                    ⏳ 사용자 승인 대기

Phase 1: 비밀번호 UX 개선
├── ...
└── PR: #{번호}
```

템플릿:
```
Phase N: {설명}
├── {세부 작업 A}  ({상태})
├── {세부 작업 B}  ({상태})
└── PR: #{번호} ({상태})
```

---

## 11. 작업 전 필수 확인 명령어 요약

```bash
# === Git 상태 ===
git log --all --oneline --graph --decorate     # 브랜치 히스토리
git branch -a                                   # 브랜치 목록
git log --oneline <branch> -5                   # 브랜치 최근 커밋
git merge-base A B                              # 공통 조상 확인

# === GitHub 상태 ===
gh pr list --state all --json number,title,state,headRefName,baseRefName
gh issue list --state all --json number,title,state,labels
gh run list --limit 10 --json databaseId,displayTitle,conclusion,status

# === 파일 시스템 ===
find . -maxdepth 3 -not -path './.git/*' -not -path '*/node_modules/*' | sort
grep -rn "target" --include="*.go"              # 함수/변수 존재 확인
```

---

## 12. 주의사항 및 금지 패턴

### 금지
- ❌ 하나의 PR에 BE + FE + DevOps + 문서 변경을 모두 섞음
- ❌ `main`에 직접 push
- ❌ PR 설명(바디) 없이 PR 생성 (`--body` 필수)
- ❌ 본문 코드 블록 안에 `Closes #N` 넣기 (GitHub 미인식)
- ❌ 브랜치 전략 우회 (feat → main 직행)
- ❌ 파일/함수/경로를 추측으로 사용 (grep으로 실제 존재 확인 필수)

### 주의
- ⚠️ 기존 브랜치가 `main` 기반일 경우 `dev`로의 머지 방안 확인 필요
- ⚠️ Phase 계획은 사용자 승인 후 실행 (예외: 명시적 지시가 있을 경우 생략 가능)
- ⚠️ 각 Phase는 이전 Phase의 머지 여부와 무관하게 설계하되, 실제 실행은 순차적
- ⚠️ PR 머지 후에는 로컬에서 `git checkout main && git pull origin main`으로 최신화
