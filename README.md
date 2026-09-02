# discord-evenotify

<img width="2000" height="744" src="https://github.com/user-attachments/assets/593e69b4-0a48-410b-8ff6-b606f0e560ca" />
DiscordのScheduled Eventの前日に、「興味あり」を押したユーザーへ個別DMでリマインドを送る常駐Botです。DB／KVストアを持たない完全ステートレス構成で、1日1回のバッチが起動中の全ギルドを走査します。

## セットアップ

### 1. Discord Developer PortalでのBot作成

1. [Discord Developer Portal](https://discord.com/developers/applications) で新しいアプリケーションを作成する。
2. 左メニューの「Bot」からBotユーザーを追加し、**Bot Token** を発行する（この値は後述の環境変数に設定する）。
3. 「OAuth2 > URL Generator」で `bot` スコープを選び、以下の権限を付与した招待URLを発行して開発用サーバーにBotを招待する。
   - Scheduled Eventsの閲覧（View Channels 等、Scheduled Events APIの利用に必要な基本権限）
   - ユーザーへのDM送信（BotとDMチャンネルを開けること。Bot自体には特別な権限設定は不要だが、対象ユーザー側でBotからのDMがブロックされていないこと）

### 2. 環境変数の設定

| 環境変数 | 必須 | 説明 |
|---|---|---|
| `EVEPING_DISCORD_TOKEN` | ✅ | 手順1で発行したBot Token |
| `EVEPING_DRY_RUN` | - | `true`または`1`を設定するとdry-runモードで起動する。対象イベント・対象ユーザーへ送るはずのDM内容をログに出力するのみで、実際のDM送信（`SendDM`）は行わない。接続確認や動作確認のためにBotを起動するだけで本物のリマインドDMが送られてしまう事態を避けるためのフラグ（詳細はGitHub Issue #19）。 |

ローカル開発では `.env` などに保存して読み込んで構わないが、`.env` はリポジトリにコミットしないこと（`.gitignore` で除外済み）。

```bash
export EVEPING_DISCORD_TOKEN="xxxxxxxx.xxxxxx.xxxxxxxxxxxxxxxxxxxxxxxx"
```

### 3. ビルド・起動

Go 1.24以降が必要（`go.mod` の `go 1.24.7` 指定に合わせる）。

```bash
go build ./cmd/eveping
./eveping
```

または `go run` でそのまま起動できる。

```bash
go run ./cmd/eveping
```

起動すると、Botセッションを開いたのち内部スケジューラが24時間ごとに日次バッチ（`RunDailyBatch`）を実行する常駐プロセスとなる。標準出力にバッチの開始・終了・対象イベント数・DM送信の成功／失敗件数がログとして出力される。

## Dockerでの起動

Goツールチェインを用意せずに、コンテナイメージとしてビルド・起動することもできる。イメージには`EVEPING_DISCORD_TOKEN`を焼き込まないため、起動時に環境変数として渡す。

### イメージのビルドと起動

```bash
docker build -t eveping .
docker run -e EVEPING_DISCORD_TOKEN="xxxxxxxx.xxxxxx.xxxxxxxxxxxxxxxxxxxxxxxx" eveping
```

接続確認だけしたい場合は `-e EVEPING_DRY_RUN=true` を追加する（`EVEPING_DRY_RUN`の詳細は上記「環境変数の設定」を参照）。

```bash
docker run -e EVEPING_DISCORD_TOKEN="xxxxxxxx.xxxxxx.xxxxxxxxxxxxxxxxxxxxxxxx" -e EVEPING_DRY_RUN=true eveping
```

### docker composeでの起動

`docker-compose.yml` を使う場合、`EVEPING_DISCORD_TOKEN`（および必要に応じて `EVEPING_DRY_RUN`）を `.env` ファイル（コミットしないこと）等で用意した上で起動する。

```bash
echo 'EVEPING_DISCORD_TOKEN=xxxxxxxx.xxxxxx.xxxxxxxxxxxxxxxxxxxxxxxx' > .env
docker compose up
```

### ベースイメージの更新

`Dockerfile` のベースイメージ（`golang`/`alpine`）はビルド再現性のためdigestで固定している。セキュリティパッチ等で更新する場合は、対象タグの最新digestを取得して `Dockerfile` の該当行を書き換えること。

## 開発用Discordサーバーでの手動検証手順

自動テストはDiscord APIとの実通信を含まないため、実際にDMが届くことは以下の手順で目視確認する。

Botは起動直後に即座に日次バッチを実行する設計のため（再起動のたびに「翌日」判定をやり直すため。詳細は`internal/scheduler`のコメントを参照）、接続確認だけしたい・DM送信はせずに済ませたい場合は `EVEPING_DRY_RUN=true` を設定して起動する。対象イベント・対象ユーザーの一覧はログに出力されるが、実際のDM送信は行われない。EvePingは二重送信対策を実装しない設計のため、通常起動での動作確認を繰り返すと同一イベントの興味ありユーザーへリマインドDMが重複して届く点に注意する。

1. 開発用Discordサーバーに、セットアップ手順どおりBotを招待する。
2. サーバー内で **開始日時が翌日（UTC基準）** のScheduled Eventを作成する。
3. 検証用アカウントでそのイベントに対して「興味あり（Interested）」を押す。
4. `EVEPING_DRY_RUN` を設定せずBotプロセスを起動し、バッチが実行されるのを待つ（本番の周期は24時間のため、動作確認時は一時的にコード側の `batchInterval` を短縮するか、`internal/batch.RunDailyBatch` を検証用スクリプトから直接呼び出して確認する）。
5. 検証用アカウントに、イベント名・開始日時・イベントリンクを含むDMが届くことを確認する。
6. コンソールログに対象イベント数・送信成功数・送信失敗数が出力されていることを確認する。

## ビルド・テストコマンド

```bash
go build ./...
go vet ./...
go test ./...
```

## アーキテクチャ

- `cmd/eveping/main.go` — エントリポイント。環境変数からBotトークンを読み込み、discordgoセッションを開始し、スケジューラを起動する。
- `internal/discordclient` — discordgoとの結合点をインターフェース化した層。本番実装（discordgoラッパー）とテスト用のインメモリFakeを提供する。`DryRunClient`は`Client`をラップし、`SendDM`をログ出力のみに差し替える（`EVEPING_DRY_RUN`用、詳細はGitHub Issue #19）。
- `internal/batch` — バッチのコアロジック（すべて `internal/discordclient.Client` を介した純粋・テスタブルな実装）。
  - `FilterTargetEvents` — 翌日（UTC）開始かつステータスがSCHEDULED/ACTIVEのイベントを抽出する純粋関数。
  - `FetchAllInterestedUsers` — 興味ありユーザーをページネーションしながら全件取得する。
  - `SendReminderDM` — 1ユーザーへのDM送信を試み、失敗してもpanicせずエラーを返す。
  - `RunDailyBatch` — 上記を組み合わせ、全ギルド・全対象イベント・全ユーザーに対してDM送信を試行し、成功／失敗件数を集計する。
- `internal/reminder` — DM本文のフォーマット（イベント名・開始日時・イベントURLを含む）。
- `internal/scheduler` — `time.Ticker` を使い、注入可能な周期でコールバックを呼び続ける常駐ループ。
- `Dockerfile` / `.dockerignore` — マルチステージビルド（Go公式イメージでビルド → 軽量な実行イメージ）によるコンテナイメージ定義。Botトークンはイメージに埋め込まず、コンテナ起動時の環境変数（`EVEPING_DISCORD_TOKEN`）としてのみ注入する。

DB・KVストアは使用せず、全ての状態はバッチ実行のたびにDiscord APIから取得する（二重送信対策は実装しない設計判断について、詳細はGitHub Issue #4のEpicを参照）。
