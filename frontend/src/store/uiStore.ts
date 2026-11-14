import { create } from 'zustand';
import type { Position } from '../types/game';
import type { UnitType } from '../types/constants';

interface UIStore {
  selectedUnitType: UnitType | null;
  selectedUnitId: string | null;
  selectedBuildingId: string | null;
  commandTarget: Position | null;
  hoveredCell: { col: number; row: number } | null;
  showUpgradePanel: boolean;
  setSelectedUnitType: (type: UnitType | null) => void;
  setSelectedUnitId: (id: string | null) => void;
  setSelectedBuildingId: (id: string | null) => void;
  setCommandTarget: (position: Position | null) => void;
  setHoveredCell: (cell: { col: number; row: number } | null) => void;
  setShowUpgradePanel: (show: boolean) => void;
}

export const useUIStore = create<UIStore>((set) => ({
  selectedUnitType: null,
  selectedUnitId: null,
  selectedBuildingId: null,
  commandTarget: null,
  hoveredCell: null,
  showUpgradePanel: false,
  setSelectedUnitType: (type) => set({ selectedUnitType: type, selectedUnitId: null, selectedBuildingId: null }),
  setSelectedUnitId: (id) => set({ selectedUnitId: id, selectedUnitType: null, selectedBuildingId: null }),
  setSelectedBuildingId: (id) => set({ selectedBuildingId: id, selectedUnitType: null, selectedUnitId: null }),
  setCommandTarget: (position) => set({ commandTarget: position }),
  setHoveredCell: (cell) => set({ hoveredCell: cell }),
  setShowUpgradePanel: (show) => set({ showUpgradePanel: show }),
}));
