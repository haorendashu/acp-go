
path: internal/cmd/schema
desc: json 스키마에서 golang 정의 파일을 생성하기 위한 cli 프로그램


## 테스트 가이드
go.mod, go.sum 파일이 내부에 있음으로 테스트 코드를 실행할때는 프로젝트 루트가 아니라 `internal/cmd/schema` 내부에서 실행해야함

- 테스트를 위해 빌드파일을 생성하지 마시오. `go run` 사용
- 아웃풋 파일을 검사해야 하는경우 출력 위치를 임시폴더에 생성
- CLI 프로그램이 정상 실행되는지 확인해야 하는경우 프로젝트 루트에서 `go generate` 실행