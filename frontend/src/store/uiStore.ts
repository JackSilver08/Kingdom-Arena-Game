import { create } from 'zustand';
import type { UnitType } from '../types/constants';

interface UIStore {
  selectedUnitType: UnitType | null;
  selectedBuildingId: string | null;
  hoveredCell: { col: number; row: number } | null;
  showUpgradePanel: boolean;
  setSelectedUnitType: (type: UnitType | null) => void;
  setSelectedBuildingId: (id: string | null) => void;
  setHoveredCell: (cell: { col: number; row: number } | null) => void;
  setShowUpgradePanel: (show: boolean) => void;
}

export const useUIStore = create<UIStore>((set) => ({
  selectedUnitType: null,
  selectedBuildingId: null,
  hoveredCell: null,
  showUpgradePanel: false,
  setSelectedUnitType: (type) => set({ selectedUnitType: type, selectedBuildingId: null }),
  setSelectedBuildingId: (id) => set({ selectedBuildingId: id, selectedUnitType: null }),
  setHoveredCell: (cell) => set({ hoveredCell: cell }),
  setShowUpgradePanel: (show) => set({ showUpgradePanel: show }),
}));
