import { GameRenderer } from './renderer';

export class InputHandler {
  private renderer: GameRenderer;
  private canvas: HTMLCanvasElement;
  private onCellClick: ((col: number, row: number) => void) | null = null;
  private onCellRightClick: ((col: number, row: number) => void) | null = null;
  private onCellHover: ((col: number, row: number) => void) | null = null;
  private onCellLeave: (() => void) | null = null;

  private handleMouseMove = (event: MouseEvent) => {
    const cell = this.renderer.getCellFromPixel(event.clientX, event.clientY);
    if (cell) {
      this.renderer.setHoveredCell(cell);
      this.onCellHover?.(cell.col, cell.row);
    } else {
      this.renderer.setHoveredCell(null);
      this.onCellLeave?.();
    }
  };

  private handleMouseLeave = () => {
    this.renderer.setHoveredCell(null);
    this.onCellLeave?.();
  };

  private handleClick = (event: MouseEvent) => {
    if (event.button !== 0) return;
    const cell = this.renderer.getCellFromPixel(event.clientX, event.clientY);
    if (cell) this.onCellClick?.(cell.col, cell.row);
  };

  private handleContextMenu = (event: MouseEvent) => {
    event.preventDefault();
    const cell = this.renderer.getCellFromPixel(event.clientX, event.clientY);
    if (cell) this.onCellRightClick?.(cell.col, cell.row);
  };

  constructor(canvas: HTMLCanvasElement, renderer: GameRenderer) {
    this.canvas = canvas;
    this.renderer = renderer;
    this.canvas.addEventListener('mousemove', this.handleMouseMove);
    this.canvas.addEventListener('mouseleave', this.handleMouseLeave);
    this.canvas.addEventListener('click', this.handleClick);
    this.canvas.addEventListener('contextmenu', this.handleContextMenu);
  }

  setOnCellClick(callback: (col: number, row: number) => void) {
    this.onCellClick = callback;
  }

  setOnCellRightClick(callback: (col: number, row: number) => void) {
    this.onCellRightClick = callback;
  }

  setOnCellHover(callback: (col: number, row: number) => void) {
    this.onCellHover = callback;
  }

  setOnCellLeave(callback: () => void) {
    this.onCellLeave = callback;
  }

  destroy() {
    this.canvas.removeEventListener('mousemove', this.handleMouseMove);
    this.canvas.removeEventListener('mouseleave', this.handleMouseLeave);
    this.canvas.removeEventListener('click', this.handleClick);
    this.canvas.removeEventListener('contextmenu', this.handleContextMenu);
  }
}
