# History Logging — 에이전트 작업 히스토리 자동 로깅

> 모든 작업 완료 시점에 반드시 준수해야 하는 히스토리 기록 정책.
> 이 정책은 2인 이상 협업 환경에서 서로 다른 AI 에이전트 간 작업 파편화를 방지하고 변경 사항을 투명하게 추적하기 위해 존재한다.

---

## 산출물 텍스트 형식

프로젝트에서 생성하는 모든 산출물은 아래 형식 규칙을 따른다:

- **불릿 포인트 필수**: 한 문단으로 나열하지 않고 불릿 리스트(`-`)로 구조화한다.
- **한 불릿 한 정보**: 각 불릿은 하나의 정보 단위만 담는다. 여러 사실을 한 줄에 나열하지 않는다.
- **중첩 허용**: 상위 항목과 하위 항목은 들여쓰기로 계층을 표현할 수 있다.
- **적용 대상**: 히스토리 로그, 커밋 메시지 본문, 계획 문서, 분석 보고, 코드 리뷰 등

---

## 목적

AI 에이전트 간의 작업 파편화를 막고 변경 사항을 투명하게 추적하기 위해, 모든 작업(기능 구현, 리팩토링, 버그 수정 등) 완료 시점에 반드시 지정된 형식으로 작업 히스토리를 기록해야 한다.

---

## 엄격한 규칙

다음 규칙은 어떠한 경우에도 절대 위반해서는 안 된다:

### 과거 히스토리 읽기 금지 (컨텍스트 절약)

- 사용자의 명시적인 요청이 없는 한, `agent/history/` 하위 경로에 있는 기존 파일들을 먼저 읽거나 탐색(`read`, `cat`, `grep`, `find` 등)해서는 절대 안 된다.
- 이는 불필요한 프롬프트 캐시 무효화 및 토큰 낭비를 원천 차단하기 위함이다.
- 히스토리는 "기록 전용"이다. 추적이 필요할 때만 명시적으로 열람한다.

### Git 충돌 방지

- 반드시 현재 Git 환경의 사용자 이름(`git config user.name`)을 기반으로 지정된 독립 디렉토리에만 파일을 생성/수정해야 한다.
- 서로 다른 사용자가 동일한 파일에 기록하여 충돌이 발생하는 것을 방지한다.

### 기록 방식 (Append-Only)

- 오늘 날짜의 파일이 없다면 새로 생성한다.
- 이미 존재한다면 기존 내용을 절대 수정하거나 삭제하지 않고, **가장 최하단에 내용을 추가(Append)** 한다.

---

## 기록 경로 및 파일명 패턴

```
agent/history/{git_username}/{branch}.md
```

| 구성요소 | 설명 |
|----------|------|
| `{git_username}` | `git config user.name` 명령어로 확인된 Git 사용자 이름 (소문자, 공백은 하이픈으로 치환) |
| `{branch}` | `git branch --show-current`로 획득한 현재 브랜치명. `/`는 `-`으로 치환 (예: `task/49` → `task-49`, `feature/user-auth` → `feature-user-auth`) |

**획득 명령어**:
```bash
git config user.name
git branch --show-current
```

**예시 경로**:
```
agent/history/inley/task-49.md
agent/history/john-doe/hotfix-timeout-fix.md
```

브랜치명을 파일명으로 사용하면 하나의 브랜치(태스크)에 하나의 히스토리 파일이 대응되어, 동시 작업 시 충돌 없이 모든 로그가 태스크 단위로 축적된다.

---

## 히스토리 로깅 템플릿

작업을 마치고 해당 날짜 파일에 로그를 기록할 때, 아래 마크다운 템플릿을 **엄격하게 준수**해야 한다.
`[...]`로 표시된 부분은 동적으로 수집하고 요약한 내용으로 대체한다.

```markdown
### 작업의 핵심 요약 제목 (사용 에이전트명)
- **User Intent**: [증상 + 요청 — 무슨 일이 일어나서, 무엇을 해달라고 했는지. 구체적 증상 기술, 1~2줄]
- **Agent Context**: [진단 + 접근 — 실제 근본 원인과 해결 방식. 사용자 요청과 다른 접근을 취했다면 그 이유, 1~2줄]
- **Key Decisions**:
  - [핵심 설계/추론 요약 1] — 왜 이 방식을 선택했는지 기술스택 컨벤션 근거와 함께 1줄로 명시
  - [핵심 설계/추론 요약 2] — 변환 방식, 영속성 전이, 성능 최적화 등 판단 근거 기술
- **Affected Files**:
  - **Created**:
    - `[프로젝트 루트 기준 상대 경로]` — 간략 설명
  - **Modified**:
    - `[프로젝트 루트 기준 상대 경로]` (+N/-M) — 변경 핵심 요약
  - **Deleted**:
    - `[프로젝트 루트 기준 상대 경로]` — 삭제 사유
```

---

## 각 필드별 작성 규칙

### 기록 시점

- 히스토리 로그는 **사용자가 에이전트에 요청한 작업 단위가 완료된 시점**에 기록한다. 컴포넌트 단위 커밋마다 기록하지 않는다.
  - 작은 작업은 단일 응답 내에서 완료될 수 있다.
  - 큰 작업은 여러 번의 응답(세션 분할)에 걸쳐 진행될 수 있다.
- 히스토리 기록은 **에이전트 응답의 가장 마지막 단계**로 수행하며, 히스토리 파일을 마지막 커밋에 포함한다.

### 작업 제목 (Title)
- 작업의 핵심을 1문장으로 압축한다. 한글로 작성하며, 너무 길지 않게 (15자 내외 권장).
- **태스크 번호는 현재 브랜치명에서 추출**한다. 브랜치가 `task/49` 또는 `feature/42`와 같이 번호를 포함하는 패턴일 경우, `[TASK-49]` 또는 `[TASK-42]`를 제목 앞에 붙인다.
- 브랜치명이 패턴과 일치하지 않거나 번호를 추출할 수 없는 경우(예: `main`, `dev`, `fix/login`, `hotfix/timeout`), **prefix를 표시하지 않고 제목만 기록**한다. `[TASK-XX]`와 같은 추측성 표기를 절대 사용해서는 안 된다.
- 브랜치 확인 명령어: `git branch --show-current`
- 예시 (태스크 브랜치 `task/49`): `[TASK-49] 재고 마스터 페이징 조회 API 구현`
- 예시 (일반 브랜치 `hotfix/timeout`): `JWT 토큰 갱신 로직 리팩토링`

### User Intent (사용자 의도)

- **증상 + 요청 구조**로 작성한다. 불릿 포인트를 사용하여 다음을 포함한다:
  - 사용자가 관찰한 구체적 증상 ("타임아웃 발생", "화이트 스크린" 등)
  - 사용자가 요청한 해결 방향
  - 작업의 배경이 되는 제약 조건이나 컨텍스트
- "무엇이 일어나서, 무엇을 해달라고 했는지"가 2주 후에 봐도 맥락이 복원될 수준이어야 한다.
- 추상적 표현("느리다", "안 된다") 대신 구체적 증상을 기록한다.
- 글자 수나 줄 수 제한은 없다. 맥락 복원에 필요한 만큼 상세히 기술한다.

### Agent Context (에이전트 맥락)

- **진단 + 접근 구조**로 작성한다. 불릿 포인트를 사용하여 다음을 포함한다:
  - 실제 근본 원인 진단
  - 선택한 해결 방식과 그 이유
  - User Intent와 실제 작업 간의 간극이 있었다면 그 차이
  - 사용자가 요청한 해결책과 다른 접근을 취했다면 그 이유
- 글자 수나 줄 수 제한은 없다.

### Key Decisions

- **"무엇을 했는지"가 아닌 **"왜 이 방식을 선택했는지"**를 불릿 포인트로 기술한다.
- 해당 도메인의 기술스택 컨벤션 문서(`agent/project/`)를 근거로 인용한다.
- 각 항목은 결정 사항 1개 + 근거 1개의 구조로 작성한다.
- 글자 수나 줄 수 제한은 없다. 판단 근거가 복잡하면 필요한 만큼 상세히 기술한다.

### Affected Files

- `<details><summary>` 토글 섹션으로 감싸서 작성한다. 파일이 많을 때 가독성을 높이기 위함이다.
- 프로젝트 루트 기준 상대 경로로 기입한다.
- 생성(Created), 수정(Modified), 삭제(Deleted)로 그룹화한다.
- 형식:
  ```markdown
  - **Affected Files**: <details><summary>N개 파일</summary>

    - **Created**:
      - `{경로}` — 간략 설명
    - **Modified**:
      - `{경로}` (+N/-M) — 변경 핵심 요약
    - **Deleted**:
      - `{경로}` — 삭제 사유

    </details>
  ```

---

## 사용 에이전트명 표기

| 에이전트 | 표기 |
|----------|------|
| Claude Code (Claude 모델) | `Claude Code` |
| Cursor (Claude 모델) | `Cursor` |
| Cursor (GPT 모델) | `Cursor (GPT)` |
| GitHub Copilot | `Copilot` |
| Codex CLI | `Codex` |
| 기타 | 실제 에이전트명 그대로 |

---

## 완전한 예시

**태스크 브랜치(`task/63`)에서 작업한 경우:**

```markdown
### [TASK-63] 재고 마스터 페이징 조회 API 구현 (Claude Code)
- **User Intent**: 재고 현황 화면에서 대량 데이터 로딩 시 타임아웃이 발생하여, 페이지네이션과 검색 필터를 적용한 API 개선 요청
- **Agent Context**: 3천 건 이상의 입고 데이터가 쌓이며 DB 풀스캔이 발생하는 것이 근본 원인으로 진단. Pageable + Specification 조합으로 동적 쿼리와 페이지네이션을 동시에 적용하여 쿼리당 20건으로 제한.
- **Key Decisions**:
  - JPQL이 아닌 Spring Data JPA `Pageable` 인터페이스 사용 — agent/backend/overview.md Repository 규칙 준수
  - 검색 조건은 QueryDSL 대신 Specification으로 구현하여 동적 쿼리 구성의 복잡도 억제
  - 프론트엔드 페이지 크기는 20으로 고정 — 모바일 대응 시나리오 고려
- **Affected Files**: <details><summary>6개 파일</summary>

  - **Created**:
    - `backend/.../global/response/PageResponse.java` — 범용 페이징 응답 DTO
  - **Modified**:
    - `backend/.../api/inventory/InventoryController.java` (+35/-12)
    - `backend/.../domain/inventory/repository/InventoryRepository.java` (+28/-5)
    - `backend/.../domain/inventory/service/InventoryServiceImpl.java` (+42/-8)
    - `frontend/src/api/inventoryApi.ts` (+15/-4)
    - `frontend/src/views/InventoryStatusView.vue` (+42/-18)

  </details>
```

**일반 브랜치(`hotfix/timeout-fix`)에서 작업한 경우:**

```markdown
### 로그인 페이지 토큰 만료 처리 개선 (Claude Code)
- **User Intent**: 토큰 만료 시 화이트 스크린이 발생하여, 만료 시점에 자동으로 로그인 페이지로 이동하도록 요청
- **Agent Context**: 401 응답을 인터셉트하지 못해 발생. 브랜치명이 `hotfix/timeout-fix`로 태스크 번호가 없어 prefix 없이 기록.
- **Key Decisions**:
  - 기존 `getAuthHeaders()` 유틸리티 함수 내에서 401 응답 감지 및 리다이렉트 처리 — 최소 변경
- **Affected Files**: <details><summary>1개 파일</summary>

  - **Modified**:
    - `frontend/src/api/authApi.ts` (+12/-3)

  </details>
```
