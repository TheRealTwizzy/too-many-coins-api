# Too Many Coins!

Too Many Coins! is an online-only massively multiplayer website game built around inflation, scarcity, and shared economic pressure. It is designed for everyone and supports many thousands of concurrent players in each season. The game is simple to understand but strategically deep over time.

---

## Overview

Players earn Coins and spend them to buy Stars. Stars determine leaderboard rank for a season. As more Coins enter the system and as time passes, Stars become increasingly expensive. Coin supply shrinks as the season progresses, creating scarcity and tension, especially near the end. Coin shortage is possible but rare; the system stays liquid enough for daily action.

The game runs in fixed-length seasons and resets regularly, while preserving long-term player progression through cosmetics, titles, badges, and history.

---

## Core Design Principles

The game must be simple, transparent, and fair  
All economy logic must be enforced server-side  
Bulk buying must be technically allowed but economically discouraged  
Late-season scarcity must feel intense but still rewarding  
Economic pressure must curve, not cliff  
There is always a rational action and a cost to inaction  
There is never a safe move  
The system must resist coordinated manipulation and bad actors  
Players must have reasons to stay active until the end of a season and return for future seasons  

---

## Seasons

Each season lasts exactly 28 calendar days.  
Up to 4 seasons may be active at the same time.  
Season start times are staggered one week apart.  
Players may join any active season at any time.  
Each season has its own independent economy.  
Coins and Stars reset at the end of each season.  
Persistent rewards carry over between seasons.

---

## Resources

Coins are the currency earned and spent by players.  
Stars are points that determine leaderboard position.  
Stars are minted only by the system; trading only transfers existing Stars.  
Players may trade with one another only through a brokered, system-controlled process under strict eligibility rules.

---

## Core Gameplay Loop

Players earn Coins through limited daily and activity-based systems.  
Players spend Coins to buy Stars.  
Players may optionally trade Coins for existing Stars under tight, time-worsening constraints.  
Star prices increase over time and with demand.  
Coin supply decreases over time.  
Inflation pressure increases monotonically; delay is punished and mistakes are permanent.  
Late-season decisions become harder and more consequential.

---

## Trading (Conditional, Brokered)

Trading is optional, costly, asymmetric, and increasingly restrictive as the season progresses.

Trading rules:

Trades are Coins-for-Stars only (no Coin-for-Coin, no Star-for-Star).  
Trades are brokered by the system; players do not set prices.  
Every trade burns Coins as overhead. Burned Coins never re-enter the economy.  
Trades never create Coins or Stars and never bypass scarcity.  
Trades are asymmetric: the buyer pays more than the seller receives due to burn and fees.  
Trades are priced at or above the current system star price, plus a time-based premium.  
Trades always contribute to market pressure, never relieve it.  

Eligibility gates (must pass all):

Both players must be currently active and time-normalized participants.  
Both must have recent coin spending activity (no pure hoarders).  
Relative Star holdings must be within a tightening ratio band.  
Coin liquidity must be within a tightening band (not too low, not too high).  
Inflation exposure difference must be within a tightening band.  

Some players will not qualify to trade. Some pairs will never qualify.

As the season progresses:

Trade eligibility gates tighten.  
Trade burn percentage rises.  
Maximum Stars per trade drops.  
Daily trade limits decrease.  

Late-season trading is expensive, dangerous, and narrow, but still rational in specific cases.

Typical rational cases: a seller needs liquidity to keep playing, or a buyer pays a premium to reach a tier when time is short.

---

## Trading as Pressure, Not Relief

Trading does not save players from inflation. It adds pressure:

Trades burn Coins and reduce total liquidity.  
Trades are priced with a premium and never undercut the system price.  
Trades increase market pressure and can make future Stars more expensive.  
Eligibility tightens over time and can deny trades entirely.  

Trading is a costly tool for repositioning risk, not a catch-up system.

---

## Star Pricing

Star prices must be dynamic and depend on multiple factors:

Time progression across the 28-day season  
Purchase quantity, with non-linear scaling for bulk purchases  
Market pressure based on recent star-buying activity  
A late-season spike that sharply increases prices in the final week  

Star prices must remain affordable relative to per-player coin emission. Prices should track average coins per player so most active players can still buy stars throughout the season.

Bulk purchases must scale so aggressively that they become almost infeasible late in the season. Bulk buying should only be viable early for players attempting to gain an early lead.

Market pressure must be smoothed using rolling averages and rate limits so prices cannot spike instantly. Coordinated manipulation on day one must be ineffective.

---

## Coin Supply and Inflation

Coins are introduced through controlled, server-managed inflation.  
Coins must never be created directly by player-triggered actions.

Coin emission rules:

Coins are emitted continuously by the server  
A global coin budget exists for each day and decreases over the season  
Players earn Coins by drawing from this pool via limited faucets  
Individual player earning caps decrease over time  
Late-season coin supply is significantly scarcer  

Coin emission must be time-sliced so a full day’s supply cannot be drained instantly. If coin consumption is too fast, earning rates must be throttled smoothly rather than stopped abruptly.

Trade burn is modeled and balanced against minting to maintain liquidity. Coin shortage is possible but rare; the system must remain liquid enough for meaningful daily action.

---

## Late-Season Design

Late-season play must feel tense but worthwhile.

Late in the season:
Star prices are much higher  
Coin supply is much lower  

Late-season incentives must not inject large amounts of Coins. Rewards should be non-economic and persistent, such as:

Badges  
Titles  
Cosmetics  
Achievements  
Participation recognition  
Community-wide progress rewards  

Late-season play also includes:

Small, high-impact star purchases  
Tight, costly trading opportunities  
Late-season challenges that reward persistence without adding Coins

---

## Anti-Manipulation and Abuse Prevention

One active player per IP address per season must be enforced by default.

If multiple accounts originate from the same IP:
They are blocked from joining a season unless whitelisted  

A whitelist request system must exist for households, schools, and shared networks.  
Whitelist approvals must be reviewed and approved manually by admins.

Players earn Coins faster while actively using the site. If the browser/tab is hidden or inactive, passive coin drip slows but does not stop.

Admin knobs:
Global: daily emission target, faucets/sinks/telemetry toggles.
Per-player: set/add coins, set/add stars, drip multiplier, pause/resume drip, touch activity.

Additional protections:
Rate-limited account creation  
CAPTCHA and verification  
Cooldown before new accounts can join a season  
Detection of suspicious clustering or coordinated behavior  
Automatic throttles for suspicious market activity  

---

## Retention Between Seasons

Players must be motivated to return season after season.

Persistent progression includes:
Account level  
Cosmetic collections  
Badges and titles  
Season history  

Each season may include a simple modifier that changes presentation or rewards without altering core economic rules.

---

## Website Pages

Landing page explaining the game quickly  
Authentication page for signup and login  
Season lobby showing all active seasons  
Main season dashboard where gameplay occurs  
Bulk purchase interface with transparent cost scaling  
Leaderboard page  
Player profile and collection page  
Settings and accessibility page  
Whitelist request page  
Internal admin console for moderation and economy monitoring  

---

## Technical Requirements

The game must scale to thousands of concurrent players per season.  
All economy calculations must be server-side only.  
Purchases must be atomic and race-condition safe.  
Real-time updates should use WebSockets or server-sent events.  
The client must never be trusted for economic logic.  
The game is online-only and web-based.

---

## Design Goal

A new player should understand the game immediately.  
An experienced player should find strategy and tension.  
Late-season play should remain meaningful.  
Players should return for multiple seasons.

---

## Mid-Season and Late-Season Play

Mid-season and late-season are designed to be risky, costly, and narrow, but never pointless.

Late joiners are disadvantaged but not invalidated.

You can still:

Earn Coins through limited faucets  
Buy Stars in small, high-impact quantities  
Trade under strict eligibility to reposition risk  
Chase tiers, badges, and late-season challenges  

You cannot:

Catch up safely  
Erase mistakes  
Avoid inflation pressure  

---

## Why You Can Still Play (Even If You Can’t Win)

You may be mathematically unable to reach first place, but you always have meaningful decisions:

Spend now vs. wait and risk higher prices  
Buy a small number of Stars vs. save for a later spike  
Sell Stars to regain liquidity vs. hold position  
Use a costly trade to reach a tier vs. accept rank decay  

There is always something at stake, and inaction always has a cost.

---

## Design Guarantee (Revised)

There is always a rational action.  
There is always something at stake.  
There is always a cost to inaction.  
There is never a safe move.  
Hope may exist; comfort must not.

---
