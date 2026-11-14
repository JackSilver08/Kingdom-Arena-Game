import type { GameState, Building, Unit } from '../types/game';
import { BOARD_COLS, BOARD_ROWS, CELL_SIZE, UnitType } from '../types/constants';

const BOARD_WIDTH = BOARD_COLS * CELL_SIZE;
const BOARD_HEIGHT = BOARD_ROWS * CELL_SIZE;

const COLORS = {
  bg: '#141424',
  gridLine: 'rgba(255, 255, 255, 0.08)',
  cellHover: 'rgba(255, 255, 255, 0.15)',
  p1Zone: 'rgba(59, 130, 246, 0.08)',
  p2Zone: 'rgba(239, 68, 68, 0.08)',
  midZone: 'rgba(234, 179, 8, 0.04)',
  borderLine: 'rgba(255, 255, 255, 0.4)',
  castle: { p1: '#2563eb', p2: '#dc2626' },
  fortress: { p1: '#4f46e5', p2: '#ea580c' },
  village: { p1: '#16a34a', p2: '#ca8a04' },
  validDeploy: 'rgba(34, 197, 94, 0.25)',
  invalidDeploy: 'rgba(239, 68, 68, 0.2)',
};

const UNIT_ICONS: Record<UnitType, string> = {
  [UnitType.SWORDSMAN]: '⚔️',
  [UnitType.ARCHER]: '🏹',
  [UnitType.CAVALRY]: '🐴',
  [UnitType.CATAPULT]: '💥',
  [UnitType.WIZARD]: '🧙',
  [UnitType.BEAST]: '🐉',
  [UnitType.DRAGON_KNIGHT]: '🛡️',
  [UnitType.REAPER]: '💀',
};

export class GameRenderer {
  private ctx: CanvasRenderingContext2D;
  private canvas: HTMLCanvasElement;
  private hoveredCell: { col: number; row: number } | null = null;
  private showDeployZone = false;
  private mySide: string = 'player1';

  constructor(canvas: HTMLCanvasElement) {
    this.canvas = canvas;
    this.ctx = canvas.getContext('2d')!;
    this.canvas.width = BOARD_WIDTH;
    this.canvas.height = BOARD_HEIGHT;
  }

  setHoveredCell(cell: { col: number; row: number } | null) {
    this.hoveredCell = cell;
  }

  setShowDeployZone(show: boolean) {
    this.showDeployZone = show;
  }

  setMySide(side: string) {
    this.mySide = side;
  }

  render(gameState: GameState) {
    this.ctx.clearRect(0, 0, BOARD_WIDTH, BOARD_HEIGHT);
    this.drawBackground();
    this.drawZones();
    this.drawGrid();
    this.drawBorderLine();
    if (this.showDeployZone) this.drawDeployZone();
    this.drawBuildings(gameState);
    this.drawUnits(gameState);
    this.drawHoveredCell();
    this.drawZoneLabels();
  }

  private drawBackground() {
    this.ctx.fillStyle = COLORS.bg;
    this.ctx.fillRect(0, 0, BOARD_WIDTH, BOARD_HEIGHT);
  }

  private drawZones() {
    // Player 1 zone (cols 0-4)
    this.ctx.fillStyle = COLORS.p1Zone;
    this.ctx.fillRect(0, 0, 5 * CELL_SIZE, BOARD_HEIGHT);
    // Neutral zone (cols 5-6)
    this.ctx.fillStyle = COLORS.midZone;
    this.ctx.fillRect(5 * CELL_SIZE, 0, 2 * CELL_SIZE, BOARD_HEIGHT);
    // Player 2 zone (cols 7-11)
    this.ctx.fillStyle = COLORS.p2Zone;
    this.ctx.fillRect(7 * CELL_SIZE, 0, 5 * CELL_SIZE, BOARD_HEIGHT);
  }

  private drawZoneLabels() {
    this.ctx.font = '11px sans-serif';
    this.ctx.fillStyle = 'rgba(59, 130, 246, 0.4)';
    this.ctx.textAlign = 'left';
    this.ctx.fillText('◄ PHE XANH (P1)', 10, 16);

    this.ctx.fillStyle = 'rgba(239, 68, 68, 0.4)';
    this.ctx.textAlign = 'right';
    this.ctx.fillText('PHE ĐỎ (P2) ►', BOARD_WIDTH - 10, 16);

    this.ctx.fillStyle = 'rgba(255, 255, 255, 0.25)';
    this.ctx.textAlign = 'center';
    this.ctx.fillText('VÙNG TRANH CHẤP', 6 * CELL_SIZE, 16);
  }

  private drawGrid() {
    this.ctx.strokeStyle = COLORS.gridLine;
    this.ctx.lineWidth = 1;
    for (let col = 0; col <= BOARD_COLS; col++) {
      this.ctx.beginPath();
      this.ctx.moveTo(col * CELL_SIZE, 0);
      this.ctx.lineTo(col * CELL_SIZE, BOARD_HEIGHT);
      this.ctx.stroke();
    }
    for (let row = 0; row <= BOARD_ROWS; row++) {
      this.ctx.beginPath();
      this.ctx.moveTo(0, row * CELL_SIZE);
      this.ctx.lineTo(BOARD_WIDTH, row * CELL_SIZE);
      this.ctx.stroke();
    }
  }

  private drawBorderLine() {
    this.ctx.strokeStyle = COLORS.borderLine;
    this.ctx.lineWidth = 2;
    this.ctx.setLineDash([8, 6]);
    const x = 6 * CELL_SIZE;
    this.ctx.beginPath();
    this.ctx.moveTo(x, 0);
    this.ctx.lineTo(x, BOARD_HEIGHT);
    this.ctx.stroke();
    this.ctx.setLineDash([]);
  }

  private drawDeployZone() {
    const cols = this.mySide === 'player1' ? [0, 1, 2, 3, 4] : [7, 8, 9, 10, 11];
    this.ctx.fillStyle = COLORS.validDeploy;
    for (const col of cols) {
      for (let row = 0; row < BOARD_ROWS; row++) {
        this.ctx.fillRect(col * CELL_SIZE, row * CELL_SIZE, CELL_SIZE, CELL_SIZE);
      }
    }
  }

  private drawBuildings(gameState: GameState) {
    for (const playerId of Object.keys(gameState.players)) {
      const player = gameState.players[playerId];
      for (const building of player.buildings) {
        this.drawBuilding(building, playerId);
      }
    }
  }

  private drawBuilding(building: Building, playerId: string) {
    const x = building.position.col * CELL_SIZE;
    const y = building.position.row * CELL_SIZE;
    const side = playerId === 'player1' ? 'p1' : 'p2';
    const padding = 6;

    let color: string;
    let icon: string;
    let title: string;
    switch (building.type) {
      case 'castle':
        color = COLORS.castle[side];
        icon = '🏰';
        title = 'Lâu Đài';
        break;
      case 'fortress':
        color = COLORS.fortress[side];
        icon = '🏯';
        title = 'Pháo Đài';
        break;
      case 'village':
        color = COLORS.village[side];
        icon = '🏘️';
        title = 'Làng';
        break;
    }

    // Card background
    this.ctx.fillStyle = color;
    this.ctx.beginPath();
    this.ctx.roundRect(x + padding, y + padding, CELL_SIZE - padding * 2, CELL_SIZE - padding * 2, 8);
    this.ctx.fill();

    // Border
    this.ctx.strokeStyle = 'rgba(255,255,255,0.6)';
    this.ctx.lineWidth = 2;
    this.ctx.stroke();

    // Icon
    this.ctx.font = '24px serif';
    this.ctx.textAlign = 'center';
    this.ctx.textBaseline = 'middle';
    this.ctx.fillText(icon, x + CELL_SIZE / 2, y + CELL_SIZE / 2 - 4);

    // Title label & Level badge
    this.ctx.fillStyle = '#ffffff';
    this.ctx.font = 'bold 9px sans-serif';
    this.ctx.fillText(`${title} Lv${building.level}`, x + CELL_SIZE / 2, y + CELL_SIZE - 12);

    // HP bar
    const hpRatio = Math.max(0, building.hp / building.maxHp);
    const barWidth = CELL_SIZE - padding * 4;
    const barHeight = 4;
    const barX = x + padding * 2;
    const barY = y + padding + 2;

    this.ctx.fillStyle = 'rgba(0,0,0,0.6)';
    this.ctx.fillRect(barX, barY, barWidth, barHeight);
    this.ctx.fillStyle = hpRatio > 0.5 ? '#22c55e' : hpRatio > 0.25 ? '#eab308' : '#ef4444';
    this.ctx.fillRect(barX, barY, barWidth * hpRatio, barHeight);
  }

  private drawUnits(gameState: GameState) {
    for (const playerId of Object.keys(gameState.players)) {
      const player = gameState.players[playerId];
      for (const unit of player.units) {
        this.drawUnit(unit, playerId);
      }
    }
  }

  private drawUnit(unit: Unit, playerId: string) {
    const x = unit.position.col * CELL_SIZE;
    const y = unit.position.row * CELL_SIZE;
    const cx = x + CELL_SIZE / 2;
    const cy = y + CELL_SIZE / 2;
    const isP1 = playerId === 'player1';

    // Unit circle
    this.ctx.beginPath();
    this.ctx.arc(cx, cy, CELL_SIZE / 2.8, 0, Math.PI * 2);
    this.ctx.fillStyle = isP1 ? '#3b82f6' : '#ef4444';
    this.ctx.fill();
    this.ctx.strokeStyle = '#ffffff';
    this.ctx.lineWidth = 2;
    this.ctx.stroke();

    // Icon
    this.ctx.font = '18px serif';
    this.ctx.textAlign = 'center';
    this.ctx.textBaseline = 'middle';
    this.ctx.fillText(UNIT_ICONS[unit.type] || '♟️', cx, cy);

    // HP bar
    const hpRatio = Math.max(0, unit.hp / unit.maxHp);
    const barW = CELL_SIZE - 20;
    this.ctx.fillStyle = 'rgba(0,0,0,0.6)';
    this.ctx.fillRect(x + 10, y + 4, barW, 3);
    this.ctx.fillStyle = hpRatio > 0.5 ? '#22c55e' : '#ef4444';
    this.ctx.fillRect(x + 10, y + 4, barW * hpRatio, 3);
  }

  private drawHoveredCell() {
    if (!this.hoveredCell) return;
    const { col, row } = this.hoveredCell;
    const validCols = this.mySide === 'player1' ? [0, 1, 2, 3, 4] : [7, 8, 9, 10, 11];
    const isValid = validCols.includes(col);

    this.ctx.fillStyle = this.showDeployZone
      ? (isValid ? COLORS.validDeploy : COLORS.invalidDeploy)
      : COLORS.cellHover;
    this.ctx.fillRect(col * CELL_SIZE, row * CELL_SIZE, CELL_SIZE, CELL_SIZE);

    this.ctx.strokeStyle = isValid ? 'rgba(34,197,94,0.8)' : 'rgba(255,255,255,0.4)';
    this.ctx.lineWidth = 2;
    this.ctx.strokeRect(col * CELL_SIZE, row * CELL_SIZE, CELL_SIZE, CELL_SIZE);
  }

  getCellFromPixel(pixelX: number, pixelY: number): { col: number; row: number } | null {
    const rect = this.canvas.getBoundingClientRect();
    const scaleX = this.canvas.width / rect.width;
    const scaleY = this.canvas.height / rect.height;
    const x = (pixelX - rect.left) * scaleX;
    const y = (pixelY - rect.top) * scaleY;
    const col = Math.floor(x / CELL_SIZE);
    const row = Math.floor(y / CELL_SIZE);
    if (col < 0 || col >= BOARD_COLS || row < 0 || row >= BOARD_ROWS) return null;
    return { col, row };
  }
}
