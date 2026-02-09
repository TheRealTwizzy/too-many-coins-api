# ADMIN-5 Bot Runner

The bot runner is a separate process that uses the same HTTP APIs as real players. It authenticates via HTTP-only session cookies, and only acts when `BOTS_ENABLED=true`.

---

## Bot Philosophy and Behavior

Bots exist to:

- Populate seasons with realistic competitive activity
- Provide baseline market activity
- Test server load and economy dynamics
- Create social presence and leaderboard competition

### Bot Learning and Behavior

Bots **learn from aggregate "normal player" behavior**:

- Bots observe patterns across many players
- Bots do NOT clone individual players
- Bots do NOT target specific players
- Bots vary in skill, strategy, and effectiveness

### Bot Skill Variation

Bots range from:

- **Cautious and inefficient** (poor timing, low activity)
- **Moderate and average** (baseline competitive behavior)
- **Aggressive and optimal** (strong timing, high activity)

This variation creates a **realistic competitive environment**.

### Bot Trading

- Bots **CAN trade with bots** (bot-to-bot trading allowed)
- Bots **CANNOT trade with players** (no bot-to-player trading)
- Bot trades contribute to market pressure and economic telemetry
- Bot trades follow the same rules as player trades

### Bot Leaderboard

Bots have **their own leaderboard**:

- Bots compete against each other
- Bots are visible on the main leaderboard with a **BOT badge**
- Bots create competitive pressure and social presence

### Bot Bad Actors

Bots **may be bad actors at realistic rates**:

- Some bots may exhibit suspicious behavior
- Anti-cheat applies **equally to bots and players**
- Bots help test anti-cheat systems
- Bots should reflect realistic abuse patterns

---

## Environment Variables

Required:
- `API_BASE_URL` — Public backend URL (e.g., https://your-app.fly.dev)
- `BOTS_ENABLED` — `true`/`false`
- `BOT_LIST` — JSON string of bot credentials and strategies **or** `BOT_LIST_PATH`

Optional:
- `BOT_LIST_PATH` — Path to a JSON file with bot configs
- `BOT_RATE_LIMIT_MIN_MS` — Minimum jitter between bots (default 3000)
- `BOT_RATE_LIMIT_MAX_MS` — Maximum jitter between bots (default 12000)
- `BOT_ACTION_PROBABILITY` — Chance to act on each bot (default 1.0)
- `BOT_MAX_ACTIONS_PER_RUN` — Max actions per run (default 1)

## Bot Config Format

```json
[
  {
    "username": "bot_alpha_01",
    "password": "LONG_RANDOM",
    "strategy": "threshold_buyer",
    "threshold": 500,
    "maxStarsPerDay": 50
  }
]
```

## Strategies

- `threshold_buyer`: buy 1 star when `currentStarPrice <= threshold` and coins >= price
- `cautious_buyer`: buy 1 star when `currentStarPrice <= coins * 0.5`
- `late_fomo`: threshold grows as the season progresses

## Local Run

```
BOTS_ENABLED=true \
API_BASE_URL=http://localhost:8080 \
BOT_LIST='[{"username":"bot_alpha_01","password":"...","strategy":"threshold_buyer","threshold":300}]' \
 go run ./cmd/bot-runner
```

## GitHub Actions

Add the following secrets:
- `API_BASE_URL`
- `BOTS_ENABLED`
- `BOT_LIST`

The workflow is defined in `.github/workflows/bot_runner.yml` and runs every 5 minutes.

## Tagging Bots

Mark bot accounts in the database (alpha).

Post‑alpha admin UI may provide Player Controls for tagging:
- Set **Bot status** to "Bot"
- Set **Bot profile** to the strategy label

Bots will appear in the leaderboard with a BOT badge.
