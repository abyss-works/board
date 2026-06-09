### 지침 문서 프로젝트 종속 언어 일반화 (Antigravity)
- **User Intent**: 현재 변경사항(stash apply로 복원된 수정 파일들) 중 특정 프로젝트/기술스택에 종속된 언어 표현을 일반화해달라는 요청
- **Agent Context**: `AGENTS.md`, `CLAUDE.md`, `agent/commit-convention.md`, `agent/history-logging.md` 4개 파일에서 종속 표현을 식별하여 일반화. 예시 섹션(완전한 예시, 예시 커밋)은 맥락 제공 목적이므로 유지하고, 규칙 서술 본문만 수정.
- **Key Decisions**:
  - 경로 예시 `backend/src/main/...`, `frontend/src/...` → `src/main/...`, `src/components/...` — 특정 디렉토리 구조에 의존하지 않는 범용 예시로 교체
  - 기술스택 괄호 주석 `(Java, Spring Boot, JPA)`, `(Vue 3, TypeScript, Pinia)`, `(Docker, Git, CI/CD)` 제거 — 링크 자체로 문서 내용 유추 가능하며, 불필요한 종속 표현 억제
  - 도메인 경계 준수 규칙의 `API, Service, Store` → `모듈` — 특정 아키텍처 계층명 대신 언어 중립적 표현 사용
  - 컴포넌트 커밋 순서의 `백엔드/프론트엔드` → `데이터 계층/클라이언트` — 기술스택 무관 계층 개념으로 대체
  - 빌드 확인 명령어 `./gradlew compileJava`, `vue-tsc --noEmit` 제거 → `agent/project/` 문서 참조 안내로 대체 — 지침 문서가 특정 도구에 고정되지 않도록
  - `history-logging.md` Key Decisions 규칙의 `기술스택 컨벤션 문서` → `해당 도메인의 기술스택 컨벤션 문서(agent/project/)` — 구체적 참조 경로 명시로 에이전트 판단 여지 제거
- **Affected Files**: <details><summary>4개 파일</summary>

  - **Modified**:
    - `AGENTS.md` (+9/-9) — 경로 예시, 기술스택 괄호, 도메인 경계 표현 일반화
    - `CLAUDE.md` (+9/-9) — AGENTS.md와 동일한 내용 동기화
    - `agent/commit-convention.md` (+8/-9) — 컴포넌트 커밋 순서 계층명 일반화, 빌드 명령어 일반화
    - `agent/history-logging.md` (+2/-2) — Key Decisions 규칙의 컨벤션 문서 참조 표현 일반화

  </details>
