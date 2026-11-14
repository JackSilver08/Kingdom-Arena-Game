// Board
export const BOARD_COLS = 12;
export const BOARD_ROWS = 10;
export const CELL_SIZE = 64; // pixels

// Players
export const PLAYER_1 = 'player1'; // Left side (Blue)
export const PLAYER_2 = 'player2'; // Right side (Red)

// Spawn zones
export const P1_SPAWN_COLS = [0, 1, 2, 3, 4];
export const P2_SPAWN_COLS = [7, 8, 9, 10, 11];

// Initial resources
export const INITIAL_GOLD = 50;
export const INITIAL_WOOD = 20;
export const INITIAL_STONE = 20;

// Game timing
export const MATCH_DURATION = 600; // 10 minutes in seconds
export const TICK_RATE = 10; // ticks per second
export const TICK_INTERVAL = 100; // ms per tick

// Building positions (col, row)
export const BUILDING_POSITIONS = {
  player1: {
    castle: { col: 0, row: 4 },
    fortress1: { col: 0, row: 1 },
    fortress2: { col: 0, row: 7 },
    village1: { col: 2, row: 1 },
    village2: { col: 2, row: 4 },
    village3: { col: 2, row: 7 },
  },
  player2: {
    castle: { col: 11, row: 4 },
    fortress1: { col: 11, row: 1 },
    fortress2: { col: 11, row: 7 },
    village1: { col: 9, row: 1 },
    village2: { col: 9, row: 4 },
    village3: { col: 9, row: 7 },
  },
};

// Village production rates per second per level
export const VILLAGE_PRODUCTION = {
  1: { gold: 2, wood: 1, stone: 1 },
  2: { gold: 4, wood: 2, stone: 2 },
  3: { gold: 6, wood: 3, stone: 3 },
};

// Building stats
export const BUILDING_STATS = {
  castle: {
    1: { hp: 500, defReduction: 0.3 },
    2: { hp: 750, defReduction: 0.4, upgradeCost: { gold: 100, wood: 50, stone: 50 } },
    3: { hp: 1000, defReduction: 0.5, upgradeCost: { gold: 200, wood: 100, stone: 100 } },
  },
  fortress: {
    1: { hp: 300, range: 2, dmg: 10 },
    2: { hp: 500, range: 3, dmg: 15, upgradeCost: { gold: 80, wood: 40, stone: 40 } },
    3: { hp: 700, range: 3, dmg: 25, upgradeCost: { gold: 160, wood: 80, stone: 80 } },
  },
  village: {
    1: { hp: 150 },
    2: { hp: 250, upgradeCost: { gold: 60, wood: 30, stone: 20 } },
    3: { hp: 400, upgradeCost: { gold: 120, wood: 60, stone: 40 } },
  },
};

// Unit types
export const UnitType = {
  SWORDSMAN: 'swordsman',
  ARCHER: 'archer',
  CAVALRY: 'cavalry',
  CATAPULT: 'catapult',
  WIZARD: 'wizard',
  BEAST: 'beast',
  DRAGON_KNIGHT: 'dragon_knight',
  REAPER: 'reaper',
} as const;

export type UnitType = (typeof UnitType)[keyof typeof UnitType];

// Unit stats
export const UNIT_STATS: Record<UnitType, {
  name: string;
  hp: number;
  atk: number;
  moveRange: number;
  attackRange: number;
  attackSpeed: number;
  trainTime: number;
  cost: { gold: number; wood?: number; stone?: number };
  unlockLevel: number;
}> = {
  [UnitType.SWORDSMAN]: { name: 'Kiếm Sĩ', hp: 80, atk: 15, moveRange: 2, attackRange: 1, attackSpeed: 1, trainTime: 3, cost: { gold: 30 }, unlockLevel: 1 },
  [UnitType.ARCHER]: { name: 'Cung Thủ', hp: 50, atk: 20, moveRange: 1, attackRange: 3, attackSpeed: 0.8, trainTime: 4, cost: { gold: 40, wood: 10 }, unlockLevel: 1 },
  [UnitType.CAVALRY]: { name: 'Kỵ Binh', hp: 100, atk: 25, moveRange: 3, attackRange: 1, attackSpeed: 1.2, trainTime: 5, cost: { gold: 60, wood: 15 }, unlockLevel: 1 },
  [UnitType.CATAPULT]: { name: 'Máy Bắn Đá', hp: 60, atk: 40, moveRange: 1, attackRange: 5, attackSpeed: 0.3, trainTime: 8, cost: { gold: 80, wood: 30, stone: 20 }, unlockLevel: 1 },
  [UnitType.WIZARD]: { name: 'Phù Thủy', hp: 40, atk: 35, moveRange: 2, attackRange: 3, attackSpeed: 0.5, trainTime: 7, cost: { gold: 80, wood: 20, stone: 20 }, unlockLevel: 2 },
  [UnitType.BEAST]: { name: 'Quái Thú', hp: 200, atk: 30, moveRange: 4, attackRange: 1, attackSpeed: 0.7, trainTime: 9, cost: { gold: 100, wood: 30, stone: 30 }, unlockLevel: 2 },
  [UnitType.DRAGON_KNIGHT]: { name: 'Hiệp Sĩ Rồng', hp: 150, atk: 40, moveRange: 2, attackRange: 2, attackSpeed: 0.8, trainTime: 12, cost: { gold: 150, wood: 50, stone: 50 }, unlockLevel: 3 },
  [UnitType.REAPER]: { name: 'Thần Chết', hp: 60, atk: 50, moveRange: 3, attackRange: 1, attackSpeed: 0.6, trainTime: 10, cost: { gold: 120, wood: 40, stone: 40 }, unlockLevel: 3 },
};
