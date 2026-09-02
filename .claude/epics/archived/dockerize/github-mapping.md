# GitHub Issue Mapping

Epic: #20 - https://github.com/Sut103/discord-evenotify/issues/20
(既存のIssue #20「Docker化する」をそのままEpic issueとして採用。新規Epic issueは作成していない)

Tasks:
- #22: マルチステージDockerfile + .dockerignore の追加 - https://github.com/Sut103/discord-evenotify/issues/22
- #23: 開発用docker-compose.ymlの追加 - https://github.com/Sut103/discord-evenotify/issues/23
- #24: READMEへのDockerビルド・起動手順の追記 - https://github.com/Sut103/discord-evenotify/issues/24
- #25: CIへのdocker buildビルド検証ステップ追加 - https://github.com/Sut103/discord-evenotify/issues/25

Synced: 2026-08-18T15:57:29Z

## Note: worktree/epic branch

CCPMの通常運用では `git worktree add ../epic-dockerize -b epic/dockerize` を作成するが、本セッションはこのIssueの作業を
`claude/issue-20-ccpm-ptdbam` ブランチのみで行うようセッション側の制約で指定されているため、別ブランチ・別worktreeは作成せず、
実装(Execute)もこのブランチ上で直接行う。
