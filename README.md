# Mention Monitor

Monitor keyword mentions across the internet and get notified via Bark. All mentions are archived to Notion for easy browsing and management.

## Features

- **11 Data Sources**: Hacker News, Reddit, GitHub, Twitter (via Nitter), Dev.to, Medium, Stack Overflow, Product Hunt, Lobsters, pkg.go.dev, Google
- **Real-time Notifications**: Push notifications via Bark (iOS)
- **Notion Integration**: All mentions stored in a Notion database for easy management
- **GitHub Actions**: Runs every 15 minutes, completely free
- **Monthly Archives**: Automatic monthly archiving with git tags

## Setup

### 1. Fork this repository

Click the "Fork" button to create your own copy.

### 2. Create Notion Database

1. Go to [Notion](https://notion.so) and create a new database
2. Add these properties:

| Property | Type | Description |
|----------|------|-------------|
| Title | Title | Mention title |
| Source | Select | hackernews, reddit, github, etc. |
| Type | Select | post, comment, issue, article, etc. |
| URL | URL | Link to original |
| Author | Text | Author name |
| Content | Text | Content excerpt |
| Keyword | Select | Matched keyword |
| Discovered | Date | When discovered |
| Status | Select | unread, read, archived |

3. Create a [Notion Integration](https://www.notion.so/my-integrations):
   - Click "New integration"
   - Give it a name (e.g., "Mention Monitor")
   - Copy the "Internal Integration Token"

4. Share the database with your integration:
   - Open your database in Notion
   - Click "..." → "Add connections"
   - Select your integration

5. Get the Database ID:
   - Open your database in Notion
   - Copy the ID from the URL: `notion.so/{workspace}/{DATABASE_ID}?v=...`

### 3. Get Bark Device Key

1. Install [Bark](https://apps.apple.com/app/bark-customed-notifications/id1403753865) on your iPhone
2. Open Bark and copy your device key from the app

### 4. Configure GitHub Secrets

Go to your fork → Settings → Secrets and variables → Actions

Add these **Secrets**:

| Secret | Description | Required |
|--------|-------------|----------|
| `NOTION_TOKEN` | Notion integration token | Yes |
| `NOTION_DATABASE_ID` | Notion database ID | Yes |
| `BARK_DEVICE_KEY` | Bark device key | Yes |
| `BARK_SERVER_URL` | Custom Bark server URL | No |
| `GH_TOKEN` | GitHub personal access token (for higher rate limits) | No |
| `GOOGLE_ALERT_URLS` | Comma-separated Google Alert RSS URLs | No |

Add this **Variable** (not secret):

| Variable | Description | Default |
|----------|-------------|---------|
| `KEYWORDS` | Comma-separated keywords to monitor | `lazypg,rebelice/lazypg` |

### 5. Enable GitHub Actions

1. Go to Actions tab in your fork
2. Click "I understand my workflows, go ahead and enable them"
3. The monitor will now run every 15 minutes

### 6. (Optional) Set up Google Alerts

For broader web monitoring:

1. Go to [Google Alerts](https://www.google.com/alerts)
2. Create an alert for your keyword
3. Change "Deliver to" to "RSS feed"
4. Copy the RSS URL
5. Add it to `GOOGLE_ALERT_URLS` secret (comma-separated if multiple)

## Data Sources

| Source | Content | Method |
|--------|---------|--------|
| Hacker News | Posts + Comments | Algolia API |
| Reddit | Posts + Comments | RSS |
| GitHub | Issues + Code imports | API |
| Twitter/X | Tweets | Nitter RSS (unstable) |
| Dev.to | Articles | API |
| Medium | Articles | RSS |
| Stack Overflow | Questions | RSS |
| Product Hunt | Products | RSS |
| Lobsters | Posts | JSON API |
| pkg.go.dev | Package imports | Web scraping |
| Google | Web pages | Google Alerts RSS |

## Manual Operations

### Trigger monitor manually

Go to Actions → Monitor Mentions → Run workflow

### Create manual archive

Go to Actions → Monthly Archive → Run workflow

You can specify a month (YYYY-MM format) or leave empty for last month.

### Download all data

```bash
git clone https://github.com/YOUR_USERNAME/mention-monitor
# All data is in data/mentions.json and data/archives/
```

## File Structure

```
mention-monitor/
├── .github/workflows/
│   ├── monitor.yml      # Runs every 15 minutes
│   └── archive.yml      # Monthly archiving
├── cmd/monitor/
│   └── main.go          # Main entry point
├── internal/
│   ├── collector/       # Data source collectors
│   ├── models/          # Data structures
│   └── notifier/        # Notion & Bark integration
├── data/
│   ├── mentions.json    # Current mentions
│   └── archives/        # Monthly archives
│       ├── 2025-01.json
│       └── ...
└── README.md
```

## Cost

**Completely free!**

- GitHub Actions: Free for public repos
- Notion API: Free
- Bark: Free (or self-host)
- All data sources: Free RSS/API

## Limitations

- Twitter/X monitoring via Nitter is unstable (instances get blocked)
- GitHub Actions may have delays during high load
- Google Alerts RSS may have a delay of a few hours

## License

MIT
