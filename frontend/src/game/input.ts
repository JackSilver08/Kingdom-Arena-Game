import { GameRenderer } from './renderer';

export class InputHandler {
  private renderer: GameRenderer;
  private canvas: HTMLCanvasElement;
  private onCellClick: ((col: number, row: number) => void) | null = null;
  private onCellHover: ((col: number, row: number) => void) | null = null;
  private onCellLeave: (() => void) | null = null;

  constructor(canvas: HTMLCanvasElement, renderer: GameRenderer) {
    this.canvas = canvas;
    this.renderer = renderer;
    this.setupListeners();
  }

  setOnCellClick(cb: (col: number, row: number) => void) { this.onCellClick = cb; }
  setOnCellHover(cb: (col: number, row: number) => void) { this.onCellHover = cb; }
  setOnCellLeave(cb: () => void) { this.onCellLeave = cb; }

  private setupListeners() {
    this.canvas.addEventListener('mousemove', (e) => {
      const cell = this.renderer.getCellFromPixel(e.clientX, e.clientY);
      if (cell) {
        this.renderer.setHoveredCell(cell);
        this.onCellHover?.(cell.col, cell.row);
      } else {
        this.renderer.setHoveredCell(null);
        this.onCellLeave?.();
      }
    });

    this.canvas.addEventListener('mouseleave', () => {
      this.renderer.setHoveredCell(null);
      this.onCellLeave?.();
    });

    this.canvas.addEventListener('click', (e) => {
      const cell = this.renderer.getCellFromPixel(e.clientX, e.clientY);
      if (cell) this.onCellClick?.(cell.col, cell.row);
    });
  }

  destroy() {
    // cleanup listeners if needed
  }
}
