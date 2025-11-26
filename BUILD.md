# Simple Uploader - 빌드 가이드

## 개발 환경 요구사항

- Go 1.21 이상
- Windows 10/11

## 프로젝트 구조

```
simple-uploader/
├── cmd/
│   ├── mounter/           # 메인 프로그램
│   │   └── main.go
│   └── mounter_debug/     # 디버그용 (콘솔 출력)
│       └── main.go
├── internal/
│   ├── config/            # 설정 파일 처리
│   │   └── config.go
│   ├── icon/              # 트레이 아이콘 생성
│   │   └── icon.go
│   └── rclone/            # rclone 관리
│       └── rclone.go
├── dist/                  # 배포 파일
│   ├── mounter.exe
│   ├── rclone.exe
│   ├── config.json
│   └── README.md
├── config.json            # 개발용 설정
├── go.mod
├── go.sum
└── BUILD.md
```

## 빌드 명령어

### 릴리스 빌드 (콘솔 창 숨김)

```cmd
cd C:\GitHub\simple-uploader
go build -ldflags="-H windowsgui" -o dist\mounter.exe .\cmd\mounter
```

### 디버그 빌드 (콘솔 출력 표시)

```cmd
cd C:\GitHub\simple-uploader
go build -o dist\mounter_debug.exe .\cmd\mounter_debug
```

### 전체 빌드

```cmd
cd C:\GitHub\simple-uploader
go build -ldflags="-H windowsgui" -o dist\mounter.exe .\cmd\mounter && go build -o dist\mounter_debug.exe .\cmd\mounter_debug
```

## 의존성 설치

```cmd
cd C:\GitHub\simple-uploader
go mod tidy
```

## 배포 파일 준비

배포 시 필요한 파일:

| 파일 | 설명 | 출처 |
|------|------|------|
| `mounter.exe` | 메인 프로그램 | 빌드 |
| `rclone.exe` | 클라우드 연결 엔진 | https://rclone.org/downloads/ |
| `config.json` | 설정 파일 | 직접 작성 |
| `README.md` | 사용 가이드 | 프로젝트 포함 |

### rclone.exe 다운로드

1. https://rclone.org/downloads/ 접속
2. Windows AMD64 버전 다운로드
3. 압축 해제 후 `rclone.exe`를 `dist/` 폴더에 복사

## 테스트

### 디버그 모드 실행

```cmd
cd C:\GitHub\simple-uploader\dist
mounter_debug.exe
```

콘솔에 연결 과정이 출력됩니다.

### 릴리스 모드 실행

```cmd
cd C:\GitHub\simple-uploader\dist
mounter.exe
```

시스템 트레이에 아이콘이 표시됩니다.

## 빌드 옵션 설명

| 옵션 | 설명 |
|------|------|
| `-ldflags="-H windowsgui"` | 콘솔 창 숨김 (GUI 모드) |
| `-o dist\mounter.exe` | 출력 파일 경로 |
| `.\cmd\mounter` | 빌드할 패키지 경로 |
