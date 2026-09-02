---
name: dockerize
status: completed
created: 2026-08-18T15:35:12Z
updated: 2026-08-19T13:47:22Z
progress: 100%
prd: .claude/prds/dockerize.md
github: https://github.com/Sut103/discord-evenotify/issues/20
---

# Epic: dockerize

## Overview

discord-evenotifyをコンテナイメージとして起動できるようにする。既存のアプリケーションロジック（`cmd/eveping`, `internal/*`）には一切手を入れず、マルチステージビルドの `Dockerfile`・`.dockerignore`・開発用 `docker-compose.yml`・README追記のみを追加する、純粋にデプロイ手段を追加するインフラ変更。

## Architecture Decisions

- **ビルドステージ**: `golang:1.24-alpine`（`go.mod` の `go 1.24.7` に対応する最新の1.24系）。依存は `discordgo`/`gorilla/websocket`/`golang.org/x/crypto`/`golang.org/x/sys` のみでCGO不要なため、`CGO_ENABLED=0` で静的バイナリをビルドする。
- **実行ステージ**: `alpine:3.20` 程度の軽量イメージ。Goツールチェインを含めず、ビルド成果物のバイナリのみをコピーする。distrolessも候補だが、初期導入ではシェルが使えて運用時のデバッグが容易なalpineを採用し、distrolessへの移行は必要になれば別タスクとする。
- **非rootユーザー**: 実行ステージ内に専用ユーザーを作成し、そのユーザーでプロセスを起動する。
- **レイヤーキャッシュ**: `go.mod`/`go.sum` を先にCOPYして `go mod download` を実行し、その後ソース全体をCOPYしてビルドすることで、依存変更のない再ビルドを高速化する。
- **秘匿情報の扱い**: `EVEPING_DISCORD_TOKEN` はDockerfile中のARG/ENVに一切埋め込まず、`docker run -e` またはCompose の `environment`/`env_file` を通じてのみコンテナに渡す。
- **ビルドコンテキスト**: `.dockerignore` を追加し、`.git`・ローカル`.env`・`.claude/` 等の不要ファイルをビルドコンテキストから除外する。

## Technical Approach

### Backend Services
既存のGoアプリケーションコード（`cmd/eveping/main.go`, `internal/*`）は変更しない。Docker化はデプロイ手段の追加のみで、アプリケーションの振る舞いに変更はない。

### Infrastructure
- `Dockerfile`（マルチステージ: build → runtime）をリポジトリルートに追加
- `.dockerignore` を追加
- 開発用 `docker-compose.yml` を追加（`.env` からのトークン注入をサポート）
- `README.md` にDockerでのビルド・起動手順を追記
- 既存の `.github/workflows/ci.yml` に `docker build .` を実行するジョブ（またはステップ）を追加し、Dockerfileのビルド失敗をPull Request時点で検知する。イメージのレジストリへのpush・公開は行わない（Out of Scope）。

## Test Strategy

**追記(実装後の方針変更)**: 以下のTest Strategyは策定当時(CCPMのTDD方針が全変更に一律適用されていた時点)のものであり、静的アサーションテスト(`dockerbuild/`パッケージ)は実際に書いて運用していた。その後CLAUDE.mdのTDD方針が「アプリケーションコード(`cmd/`・`internal/`配下)の変更に限る」とスコープ限定され、Dockerfile/docker-compose.yml/README/CI設定はいずれも対象外と判明したため、シンプルさを優先して `dockerbuild/` パッケージ(4テストファイル・11テスト)は削除した。各タスクファイルのTest Planに実施記録として残している。以降の検証はCode Review(目視)とCIの `docker build` ステップの成否のみに拠る。

### Test Types & Tools(策定当時)
- **静的アサーションテスト（自動・`go test`実行）**: 新設のテストファイルから、リポジトリルートの `Dockerfile` / `docker-compose.yml` をテキストとして読み込み、以下を機械的に検証する。
  - `Dockerfile` に複数の `FROM` 命令があること（マルチステージビルドになっていること）
  - `Dockerfile` 内に `EVEPING_DISCORD_TOKEN` の値がハードコードされていないこと（`ENV EVEPING_DISCORD_TOKEN=<値>` のような記述が存在しないこと）
  - `Dockerfile` に `CGO_ENABLED=0` が設定されていること
  - `Dockerfile` が非rootユーザーで実行される設定（`USER` 命令）を含むこと
  - `docker-compose.yml` が `EVEPING_DISCORD_TOKEN` を環境変数経由（`environment`/`env_file`）で参照しており、値をハードコードしていないこと
  - これらのテストはこの環境（Dockerデーモン不在のサンドボックス）でも `go test ./...` の一部として実行でき、CIでも同様に実行可能。
- **手動スモークテスト（Dockerデーモンが利用可能な環境でのみ実施、README/タスクのDefinition of Doneに手順を明記）**:
  - `docker build .` が成功すること
  - `docker run -e EVEPING_DISCORD_TOKEN=<有効なトークン> <image>` が `go run ./cmd/eveping` と同等に動作し、Discordへ接続してスケジューラが起動すること（既存READMEの手動検証手順に準拠）
  - `EVEPING_DISCORD_TOKEN` を指定せずに起動した場合、`main.go` の既存仕様どおり即座にエラー終了すること
  - `docker compose up`（または `docker compose config`）でCompose定義が問題なく解釈され、`.env` からトークンが注入されること
- **CI（GitHub Actions、自動）**: `.github/workflows/ci.yml` に `docker build .`（タグ付け・レジストリへのpushはしない）を実行するステップ/ジョブを追加する。既存の `test` ジョブと同様、push/pull_request両方のトリガーで実行し、ビルド失敗時はCIを赤くする。GitHub ActionsのUbuntuランナーにはDockerデーモンが標準で利用可能なため、この観点はこのサンドボックス環境の制約（daemon不在）を受けず自動化できる。

### Coverage Expectations
- PRDの各Functional Requirementは、上記の自動静的テストまたは手動検証手順のいずれかで必ずカバーする。
- 既存の `go build ./...` / `go vet ./...` / `go test ./...` が新規ファイル追加後も引き続き全て成功すること（回帰がないこと）。

### TDD Notes
- 各タスクは、対応する静的アサーションテストを先に書き（例: 対象ファイルが存在しない/条件を満たさないために失敗する状態を確認 = red）、その後 `Dockerfile`/`docker-compose.yml`/README側を実装してテストをgreenにする。
- Dockerデーモンが利用できないサンドボックス環境では実際のビルド・起動は自動テストできないため、その検証は各タスクの Definition of Done に手動検証ステップとして明記し、自動テストの対象外であることを明示する（既存README「開発用Discordサーバーでの手動検証手順」と同じ考え方）。

## Implementation Strategy

タスク数が少なく依存関係も単純なため、逐次作成する（Small epic, <5 tasks）。Dockerfile本体の確定が他タスクの前提になるため、まずDockerfile/.dockerignoreを完成させ、その後docker-compose.yml・README追記・CIジョブ追加を進める。docker-compose.yml・README追記・CIジョブ追加はDockerfile確定後は互いに独立しており並行可能（ただしCIジョブはDockerfileが実際にビルドできる状態でないと検証できないため、実質的にはTask 1の後で着手する）。

## Task Breakdown Preview

- [ ] Task 1: マルチステージ `Dockerfile` + `.dockerignore` の追加（静的アサーションテスト含む）
- [ ] Task 2: 開発用 `docker-compose.yml` の追加（`.env` からのトークン注入、静的アサーションテスト含む）
- [ ] Task 3: README へのDockerビルド・起動手順の追記
- [ ] Task 4: CI（`.github/workflows/ci.yml`）への `docker build` 検証ステップ追加

## Dependencies

- Task 2・Task 3・Task 4 は Task 1（Dockerfileの存在・仕様確定）に依存する。
- 外部依存: Dockerデーモンでのビルド・起動確認（手動検証）にはDocker環境が必要。Task 4のCI検証自体はGitHub Actionsランナー上のDockerデーモンで完結する。

## Success Criteria (Technical)

- `go build ./...` / `go vet ./...` / `go test ./...` が全て成功する。
- （Docker環境がある場合の手動検証）`docker build .` が成功し、`docker run -e EVEPING_DISCORD_TOKEN=...` で `go run ./cmd/eveping` と同等にBotが起動する。
- README の記述に従うだけでDocker起動ができる。
- CI上で `docker build .` が実行され、Pull Request時点でDockerfileのビルド失敗を検知できる。

## Estimated Effort

- Size: S
- Hours: 5-7

## Tasks Created

- [x] #22 - マルチステージDockerfile + .dockerignore の追加 (parallel: true)
- [x] #23 - 開発用docker-compose.ymlの追加 (parallel: true, depends_on: #22)
- [x] #24 - READMEへのDockerビルド・起動手順の追記 (parallel: true, depends_on: #22, #23)
- [x] #25 - CIへのdocker buildビルド検証ステップ追加 (parallel: true, depends_on: #22)

Total tasks: 4
Parallel tasks: 4 (#22完了後、#23・#24・#25は並行着手可能)
Sequential tasks: 0
Estimated total effort: 5.5 hours

## Execution Summary

全4タスクをTDD(red→green)で実装し、各タスク完了時に code-review スキル(medium)を実施。
- #22実装時にcode-reviewで検出: alpineランタイムに `ca-certificates` が無くDiscordへのTLS接続が失敗する重大な不具合を修正済み。あわせてCLAUDE.md/READMEのアーキテクチャ節を更新。
- #23・#24・#25のcode-reviewでは指摘なし。

**追記(簡素化)**: CLAUDE.mdのTDD方針スコープ限定(アプリケーションコードのみ)を受け、実装当時に書いていた静的アサーションテスト(`dockerbuild/`パッケージ、4ファイル・11テスト)は削除した。理由・実施記録は各タスクファイル・Test Strategy節参照。`go build ./...` / `go vet ./...` / `go test ./...` は削除後も全て成功。

**訂正**: 当初「この実行環境にDockerデーモンが無い」と記載していたが誤りだった。`dockerd` バイナリは実装セッションの環境に存在し、手動起動できることを確認した。ただし起動したdockerdで実際に `docker build .` を試みたところ、サンドボックス固有のネットワーク制約(アウトバウンドHTTPSを透過的にインターセプトするプロキシがビルドコンテナ内から信頼されずAlpineの `apk add ca-certificates` がTLS検証エラーになる、およびDocker Hub匿名pullのレート制限)によりローカルでは成功しなかった。一方、GitHub Actions CI(サンドボックス外の通常のネットワーク環境)では `docker build .` ステップが実際に成功している([確認済みの実行](https://github.com/Sut103/discord-evenotify/actions/runs/32193894776/job/95893861685)、20秒で完了) — これがDockerfile自体の正当性を示す実際の検証結果である。`docker run` によるDiscordへの実接続確認(実トークンが必要)と `docker compose up` の実起動確認は未実施のまま。`docker compose config` の静的パースはこの環境で確認済み。
