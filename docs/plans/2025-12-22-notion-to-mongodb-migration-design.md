# Notion to MongoDB Migration Design

## Overview

Replace Notion storage with MongoDB Cloud Atlas for mention storage.

## Changes

### Architecture

**Before:**
```
Collectors → mentions → Notion (storage) + Bark (notification)
                     ↓
            data/mentions.json (local backup)
```

**After:**
```
Collectors → mentions → MongoDB (storage) + Bark (notification)
                     ↓
            data/mentions.json (local backup, kept)
```

### MongoDB Data Structure

- Database: `mention_monitor`
- Collection: `mentions`

**Document schema:**
```javascript
{
  _id: ObjectId,
  id: "github_3746245504",
  source: "github",
  type: "issue",
  keyword: "lazypg",
  title: "...",
  content: "...",
  url: "https://...",
  author: "rebelice",
  discovered_at: ISODate,
  published_at: ISODate,
  status: "unread",
  created_at: ISODate
}
```

**Indexes:**
- `{ id: 1 }` - unique, for deduplication
- `{ url: 1 }` - backup deduplication
- `{ discovered_at: -1 }` - time-based queries

### File Changes

1. **New:** `internal/notifier/mongodb.go` - MongoDB client
2. **Modify:** `cmd/monitor/main.go` - Replace Notion with MongoDB
3. **Modify:** `go.mod` - Add `go.mongodb.org/mongo-driver/v2 v2.0.0`
4. **Delete:** `internal/notifier/notion.go`
5. **Modify:** `README.md` - Update configuration docs

### Environment Variables

**Remove:**
- `NOTION_TOKEN`
- `NOTION_DATABASE_ID`

**Add:**
- `MONGODB_URI` - Atlas connection string

## Implementation Tasks

1. Add MongoDB driver dependency
2. Create `internal/notifier/mongodb.go`
3. Update `cmd/monitor/main.go` to use MongoDB
4. Delete `internal/notifier/notion.go`
5. Update README.md
