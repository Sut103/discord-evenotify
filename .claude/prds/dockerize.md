---
name: dockerize
description: discord-evenotifyをDocker化し、コンテナイメージとして起動できるようにする
status: backlog
created: 2026-08-18T15:35:12Z
---

# PRD: dockerize

## Executive Summary

discord-evenotify（Go + discordgo製の常駐Bot）は現状、ビルド済みバイナリを直接実行する運用のみを想定している（VPS等で `go build` 済みバイナリを常駐させる）。本PRDでは、discord-evenotifyをコンテナイメージとして配布・起動できるようにする。マルチステージビルドの `Dockerfile` を追加し、Botトークンはイメージにベイクせず環境変数で注入する。GitHub Issue #20 に対応する。

## Problem Statement

- 現状はホストにGoツールチェインを用意してビルドするか、ビルド済みバイナリを個別に配布する必要があり、実行環境の再現性が低い。
- コンテナベースの運用（Docker Compose、各種PaaS、将来的なオーケストレーション基盤）に載せる手段がない。
- Epic #4（eve-ping）では「CI/CDやコンテナ化はスコープ外（必要になれば別途）」とされており、本PRDはその積み残しに対応する。

## User Stories

### 運用者として
- コンテナイメージをビルド（または取得）し、`EVEPING_DISCORD_TOKEN` 環境変数を渡して `docker run` するだけでBotを起動したい。
  - 受け入れ基準: イメージをビルドし、環境変数を設定して起動すると、既存バイナリ起動時と同じログ（バッチ開始・終了・成功/失敗件数等）が標準出力に出る。
  - 受け入れ基準: `EVEPING_DISCORD_TOKEN` を渡さずに起動した場合、`main.go` の既存仕様どおりエラー終了する。

### 開発者として
- ローカルで `docker compose up` 相当のコマンドにより、ビルド〜起動を一発で行い動作確認したい。
  - 受け入れ基準: `docker-compose.yml`（または同等の設定）を使って起動でき、`.env` 等からトークンを注入できる。

### 新規コントリビューターとして
- READMEを読むだけでDocker起動手順がわかるようにしたい。
  - 受け入れ基準: README にDockerでのビルド・起動手順が明記されている。

## Functional Requirements

1. マルチステージビルドの `Dockerfile` をリポジトリルートに追加する。
   - ビルドステージ: Go公式イメージ（`go.mod` の `go 1.24.7` に合わせたバージョン）で `cmd/eveping` をビルドする。
   - 実行ステージ: 最小イメージ（distroless もしくは alpine 等、詳細はEpicで決定）にビルド成果物のみをコピーする。
2. Botトークン（`EVEPING_DISCORD_TOKEN`）はビルド時にイメージへ埋め込まず、コンテナ起動時の環境変数として注入できること。
3. `docker build` で単体イメージがビルドできること。
4. `docker run -e EVEPING_DISCORD_TOKEN=... <image>` でBotプロセスが起動し、既存の `go run ./cmd/eveping` と同等に動作すること。
5. 開発用の `docker-compose.yml` を追加し、環境変数（`.env` 経由等）でトークンを注入して起動できるようにする。
6. README にDockerでのビルド・起動手順を追記する。
7. 既存のCI（`.github/workflows/ci.yml`）に、`docker build` によるイメージビルド検証ジョブを追加する。Pull Request・push時にDockerfileの構文崩れやビルド失敗を検知できるようにする（レジストリへのpush・公開は行わない）。

## Non-Functional Requirements

- イメージサイズ: マルチステージビルドにより、実行ステージにGoツールチェインを含めない（不要な肥大化を避ける）。
- セキュリティ: Botトークンなどの秘匿情報をイメージのレイヤーに焼き込まない。コンテナは非rootユーザーで実行することが望ましい。
- 既存の `go build ./...` / `go vet ./...` / `go test ./...` によるビルド・テストフローには影響を与えない（Docker関連ファイルの追加のみ）。

## Success Criteria

- `docker build .` がエラーなく成功する。
- ビルドしたイメージを `EVEPING_DISCORD_TOKEN` を指定して起動すると、Botとして正常に稼働する（Discordへの接続・スケジューラ起動ログが確認できる）ことを手動検証で確認する。
- `docker-compose.yml` を使ったローカル起動が可能。
- README の記述に従うだけで、初見のユーザーがDockerでBotを起動できる。
- CI上で `docker build` が実行され、Dockerfileのビルド失敗がPull Requestの時点で検知できる。

## Constraints & Assumptions

- 言語・ツールチェインは `go.mod` の `go 1.24.7` を前提とする。
- 依存パッケージはCGOを要しない（`discordgo`, `gorilla/websocket`, `golang.org/x/crypto`, `golang.org/x/sys` のみ）ため、`CGO_ENABLED=0` での静的ビルドを想定する。
- ベースイメージの具体的な選定（distroless/alpine等）やタグ運用、CI連携（イメージのビルド・プッシュ自動化）の詳細はEpic/タスク化の際に検討する。
- 永続化ボリュームは不要（アプリケーションが完全ステートレスであるため）。

## Out of Scope

- コンテナレジストリへのイメージの自動push・公開（CI上での `docker build` によるビルド検証自体は本PRDのスコープに含む。上記Functional Requirements参照）
- Kubernetes等のオーケストレーション用マニフェスト
- マルチアーキテクチャ（arm64等）ビルド対応
- 二重送信対策やDBなど、Docker化と無関係な既存スコープ外事項（Epic #4で既に対応外と合意済み）

## Dependencies

- 既存の `cmd/eveping` ビルド成果物・環境変数仕様（`EVEPING_DISCORD_TOKEN`）に変更はない前提。
- 外部依存: DockerホストでのGoベースイメージ取得（インターネットアクセス）。
