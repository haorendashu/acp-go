![agent client protocol golang banner](./imgs/banner-dark.jpg)

# Agent Client Protocol - Go 구현체

Agent Client Protocol (ACP)의 Go 구현체입니다. ACP는 _코드 에디터_(소스 코드를 보고 편집하는 대화형 프로그램)와 _코딩 에이전트_(생성형 AI를 사용하여 자율적으로 코드를 수정하는 프로그램) 간의 통신을 표준화합니다.

이것은 Go로 작성된 ACP 사양의 **비공식** 구현체입니다. 공식 프로토콜 사양과 참조 구현체는 [공식 저장소](https://github.com/zed-industries/agent-client-protocol)에서 찾을 수 있습니다.

> ⚠️ **주의**: Agent Client Protocol은 활발히 개발 중입니다. 이 구현체는 최신 사양 변경사항을 따라가지 못할 수 있습니다. 가장 최신의 프로토콜 사양은 [공식 저장소](https://github.com/zed-industries/agent-client-protocol)를 참조해 주세요.

프로토콜에 대한 자세한 내용은 [agentclientprotocol.com](https://agentclientprotocol.com/)에서 확인하세요.

## 기능

- ✅ **JSON-RPC 2.0** stdio 통신
- ✅ **에이전트 측 연결** 코딩 에이전트 구현을 위한 기능
- ✅ **클라이언트 인터페이스** 권한 요청 및 파일 작업
- ✅ **세션 관리** 다중 동시 세션 지원
- ✅ **파일 시스템 작업** (텍스트 파일 읽기/쓰기)
- ✅ **터미널 작업** (생성, 출력, 대기, 종료, 해제)
- ✅ **양방향 통신** 에이전트와 클라이언트 간
- ✅ **컨텍스트 기반 요청 처리** 적절한 취소 지원
- ✅ **미들웨어 지원** 요청/응답 처리

## 설치

```bash
go get github.com/ironpark/go-acp
```

## 빠른 시작

### 간단한 에이전트 생성

```go
package main

import (
    "context"
    "os"
    
    "github.com/ironpark/go-acp"
)

type MyAgent struct{}

func (a *MyAgent) Initialize(ctx context.Context, params *acp.InitializeRequest) (*acp.InitializeResponse, error) {
    return &acp.InitializeResponse{
        ProtocolVersion: params.ProtocolVersion,
        AgentCapabilities: acp.AgentCapabilities{
            LoadSession: false,
            McpCapabilities: &acp.McpCapabilities{
                Http: false,
                Sse:  false,
            },
            PromptCapabilities: &acp.PromptCapabilities{
                Audio:           false,
                EmbeddedContext: false,
                Image:           false,
            },
        },
    }, nil
}

func (a *MyAgent) Authenticate(ctx context.Context, params *acp.AuthenticateRequest) error {
    return nil // 인증 불필요
}

func (a *MyAgent) NewSession(ctx context.Context, params *acp.NewSessionRequest) (*acp.NewSessionResponse, error) {
    return &acp.NewSessionResponse{
        SessionId: "session-123",
    }, nil
}

func (a *MyAgent) Prompt(ctx context.Context, params *acp.PromptRequest) (*acp.PromptResponse, error) {
    // 여기서 사용자 프롬프트 처리
    return &acp.PromptResponse{
        StopReason: "end_turn",
    }, nil
}

// 다른 필수 Agent 인터페이스 메서드들 구현...

func main() {
    agent := &MyAgent{}
    
    // stdin/stdout을 사용한 에이전트 측 연결 생성
    conn := acp.NewAgentSideConnection(agent, os.Stdin, os.Stdout)
    
    // 연결 시작
    if err := conn.Start(context.Background()); err != nil {
        panic(err)
    }
}
```

## 아키텍처

이 구현체는 관심사의 명확한 분리를 제공합니다:

- **`Conn`**: stdio 통신을 처리하는 저수준 전송 계층
- **`Server`**: 에이전트 측을 위한 JSON-RPC 2.0 프로토콜 구현
- **`Client`**: 클라이언트 측 요청을 위한 JSON-RPC 2.0 프로토콜 구현
- **`AgentSideConnection`**: 에이전트 구현을 위한 고수준 ACP 인터페이스
- **`ClientSideConnection`**: 클라이언트 요청을 위한 고수준 ACP 인터페이스

## 프로토콜 지원

이 구현체는 다음 기능을 가진 ACP 프로토콜 버전 1을 지원합니다:

### 에이전트 메서드 (클라이언트 → 에이전트)
- `initialize` - 에이전트 초기화 및 기능 협상
- `authenticate` - 에이전트 인증 (선택사항)
- `session/new` - 새로운 대화 세션 생성
- `session/load` - 기존 세션 로드 (지원하는 경우)
- `session/set_mode` - 세션 모드 변경 (불안정)
- `session/prompt` - 사용자 프롬프트를 에이전트에 전송
- `session/cancel` - 진행 중인 작업 취소

### 클라이언트 메서드 (에이전트 → 클라이언트)
- `session/update` - 세션 업데이트 전송 (알림)
- `session/request_permission` - 작업에 대한 사용자 권한 요청
- `fs/read_text_file` - 클라이언트 파일시스템에서 텍스트 파일 읽기
- `fs/write_text_file` - 클라이언트 파일시스템에 텍스트 파일 쓰기
- `terminal/create` - 터미널 세션 생성 (불안정)
- `terminal/output` - 터미널 출력 가져오기 (불안정)
- `terminal/wait_for_exit` - 터미널 종료 대기 (불안정)
- `terminal/kill` - 터미널 프로세스 종료 (불안정)
- `terminal/release` - 터미널 핸들 해제 (불안정)

## 예제

완전한 작동 예제는 [example/simple_agent](../example/simple_agent) 디렉토리를 참조하세요.

## 개발

### 빌드

```bash
go build ./...
```

### 테스트

```bash
go test ./...
```

### 코드 생성

타입은 공식 JSON 스키마에서 생성됩니다:

```bash
# 공식 저장소에서 스키마 파일 업데이트
# 그다음 타입 재생성 (구현체별)
```

## 기여하기

이것은 비공식 구현체입니다. 프로토콜 사양 변경사항은 [공식 저장소](https://github.com/zed-industries/agent-client-protocol)에 기여해 주세요.

Go 구현체 이슈 및 개선사항은 이슈를 열거나 풀 리퀘스트를 보내주세요.

## 라이선스

이 구현체는 공식 ACP 사양과 동일한 라이선스를 따릅니다.

## 관련 프로젝트

- **공식 ACP 저장소**: [zed-industries/agent-client-protocol](https://github.com/zed-industries/agent-client-protocol)
- **Rust 구현체**: 공식 저장소의 일부
- **프로토콜 문서**: [agentclientprotocol.com](https://agentclientprotocol.com/)

### ACP를 지원하는 에디터

- [Zed](https://zed.dev/docs/ai/external-agents)
- [neovim](https://neovim.io) - [CodeCompanion](https://github.com/olimorris/codecompanion.nvim) 플러그인을 통해
- [yetone/avante.nvim](https://github.com/yetone/avante.nvim): Cursor AI IDE의 동작을 에뮬레이트하도록 설계된 Neovim 플러그인