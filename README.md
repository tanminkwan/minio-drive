# MinIO Drive

MinIO 오브젝트 스토리지를 Windows 네트워크 드라이브로 마운트하는 시스템 트레이 애플리케이션

## 특징

- **WebDAV 모드**: 별도 드라이버 설치 불필요 (기본)
- **WinFsp 모드**: 로컬 드라이브처럼 사용 (WinFsp 설치 필요)
- 시스템 트레이에서 간편하게 제어
- 자동 시작 및 자동 드라이브 연결 지원
- 탐색기에서 드래그앤드롭으로 파일 업로드/다운로드

## 요구사항

- Windows 10/11
- [rclone](https://rclone.org/downloads/) (배포 시 함께 포함)
- (WinFsp 모드만) [WinFsp](https://winfsp.dev/rel/)

## 빌드

```cmd
# 릴리스 빌드
go build -ldflags="-H windowsgui" -o dist\mounter.exe .\cmd\mounter

# 디버그 빌드
go build -o dist\mounter_debug.exe .\cmd\mounter_debug
```

## 설치

1. `dist` 폴더에 다음 파일 배치:
   - `mounter.exe` (빌드 결과물)
   - `rclone.exe` ([다운로드](https://rclone.org/downloads/))
   - `config.json` (설정 파일)

2. `config.json` 설정:

```json
{
  "minio": {
    "endpoint": "your-minio-server:9000",
    "access_key": "your-access-key",
    "secret_key": "your-secret-key",
    "bucket": "your-bucket",
    "use_ssl": false
  },
  "mount": {
    "type": "webdav",
    "port": 20080,
    "drive_letter": "Z",
    "auto_start": true
  }
}
```

## 설정 옵션

### minio

| 항목 | 설명 |
|------|------|
| `endpoint` | MinIO 서버 주소:포트 |
| `access_key` | Access Key |
| `secret_key` | Secret Key |
| `bucket` | Bucket 이름 |
| `use_ssl` | HTTPS 사용 여부 |

### mount

| 항목 | 설명 |
|------|------|
| `type` | `webdav` (기본) 또는 `winfsp` |
| `port` | WebDAV 서버 포트 (WebDAV 모드만) |
| `drive_letter` | 드라이브 문자 |
| `auto_start` | 시작 시 자동 연결 |

## 마운트 모드 비교

| | WebDAV | WinFsp |
|---|--------|--------|
| 드라이버 설치 | 불필요 | 필요 |
| 드라이브 유형 | 네트워크 드라이브 | 로컬 드라이브 |
| 호환성 | 일부 제한 | 완전 호환 |
| 보안 정책 | 대부분 허용 | 차단될 수 있음 |

## 프록시 환경

프록시 환경에서는 `NO_PROXY` 환경변수 설정 필요:

```
NO_PROXY=localhost,127.0.0.1,{MinIO서버IP}
```

## 프로젝트 구조

```
minio-drive/
├── cmd/
│   ├── mounter/           # 메인 프로그램 (GUI)
│   └── mounter_debug/     # 디버그 버전 (콘솔)
├── internal/
│   ├── config/            # 설정 파일 처리
│   ├── icon/              # 트레이 아이콘
│   └── rclone/            # rclone 관리
├── go.mod
├── go.sum
└── README.md
```

## 라이선스

MIT License
