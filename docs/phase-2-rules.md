# Kingdom Arena — Phase 2 Core Rules Baseline

Status: implementation baseline for the first server-side simulation slice.

## Simulation
- Fixed timestep: 10 ticks per second (100ms per tick).
- Economy and match timer update every 10 ticks.
- Unit movement opportunities are evaluated every 5 ticks.
- The game engine owns all authoritative state transitions.
- The room wrapper serializes command application, tick advancement and snapshots. A future room loop must remain the sole mutator of an Engine instance.
- Player movement intents are collected from one pre-movement state, then committed together.
- Unit attacks, cavalry impacts, Fortress attacks and Dragon Knight breath damage are applied in batches to avoid deterministic first-player lethal-hit advantages.

## Board and lanes
- Board: 12 columns x 10 rows.
- Player 1 deploys in columns 0–4; Player 2 deploys in columns 7–11.
- Top lane: rows 0–2; Mid lane: rows 3–6; Bottom lane: rows 7–9.
- Units retain their deployment lane in this initial ruleset. Lane switching is deferred.
- If an enemy unit is ahead in the same lane, a unit attempts to reach a legal attack position before continuing toward its building objective.
- Every lane must clear its enemy Village on the approach first.
- Top and Bottom lanes then target the corresponding enemy Fortress. They hold their lane after the Fortress is destroyed; lane switching is deferred.
- Mid lane then targets the enemy Castle.

## Unit movement profiles
- Swordsman: up to 2 cells forward, with a 1-cell fallback when the longer move is blocked.
- Archer: 1 cell in any of 8 directions.
- Cavalry: chess-knight L movement.
- Catapult: 1 orthogonal cell.
- Wizard: 2 diagonal cells.
- Beast: 4 cells in a straight line.
- Dragon Knight: 2 cells in any of 8 directions.
- Reaper: 3 cells forward; intermediate unit occupancy does not block the move.

Movement is lane-constrained in this baseline. Multi-cell intermediate-cell collision behavior is intentionally conservative and must be expanded in the next movement review.

## Economy and unlocks
- Each living Village produces 2 Gold, 1 Wood and 1 Stone per second at level 1.
- Level 2 produces 4/2/2 per second; level 3 produces 6/3/3.
- Unit costs, build costs and unlock levels follow the approved v1.0 plan.
- Unit train time is used as a per-player, per-unit-type deployment cooldown in this implementation slice.
- Fortress level 2 unlocks Wizard and Beast; level 3 unlocks Dragon Knight and Reaper.
- Castle level 2 grants +5% ATK; level 3 grants +10% ATK and +10% HP.

## Combat baseline
- Unit attacks are evaluated server-side and damage is resolved simultaneously per combat phase.
- Archer attacks along a clear row or column up to 3 cells.
- Catapult attacks at Chebyshev distance 3–5 and deals double damage to buildings.
- Catapult and Wizard use a 3x3 area effect around the selected unit or building target and exclude friendly units. Catapult deals double damage to buildings in the area.
- Cavalry deals immediate impact damage when it lands adjacent to an enemy unit.
- Beast receives +20 damage after a movement of at least 3 cells and attempts a one-cell knockback; Dragon Knight is immune.
- Reaper cannot be selected as a target during its first 3 seconds, triples its first attack, and takes 20% additional damage after being revealed.
- Dragon Knight performs a 3-cell horizontal breath for 20 damage every 10 seconds.
- Fortress range/damage: level 1 = 2/10, level 2 = 3/15, level 3 = 3/25.
- Castle destruction wins; simultaneous Castle destruction is a draw; timeout compares remaining Castle HP and equal HP is a draw.

## Explicitly deferred
- Lane switching and terrain modifiers.
- Terrain modifiers, lane switching and advanced pathfinding/collision resolution.
- Complete visual timing and edge-case balancing for all special abilities.
- Wiring the browser client to the authoritative engine through WebSocket (network integration milestone).
- WebSocket multiplayer room lifecycle, matchmaking and reconnect lifecycle (Phase 3).
- ELO matchmaking, replay playback UI and public ranked service (Phase 3+).

This document records the implemented initial rules baseline, not the complete final game design.

## Phase 2 implementation boundary

The Go engine and room wrapper can be exercised independently through tests and the bot simulation command. This phase does not yet connect the browser prototype to the engine over WebSocket. Until that integration exists, the current frontend prototype must not be treated as an authoritative competitive client.