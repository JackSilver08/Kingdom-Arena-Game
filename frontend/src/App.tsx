import { useEffect, useRef, useState } from 'react';
import { GameRenderer } from './game/renderer';
import { InputHandler } from './game/input';
import { useGameStore } from './store/gameStore';
import { useUIStore } from './store/uiStore';
import { ResourceBar } from './components/HUD/ResourceBar';
import { UnitPanel } from './components/UnitPanel/UnitPanel';
import {
  VILLAGE_PRODUCTION,
  UNIT_STATS,
  BUILDING_STATS,
  P1_SPAWN_COLS,
  P2_SPAWN_COLS,
} from './types/constants';

function App() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const rendererRef = useRef<GameRenderer | null>(null);
  const inputRef = useRef<InputHandler | null>(null);
  const animFrameRef = useRef<number>(0);

  const initGame = useGameStore((s) => s.initGame);
  const gameState = useGameStore((s) => s.gameState);
  const mySide = useGameStore((s) => s.mySide);
  const updateResources = useGameStore((s) => s.updateResources);
  const addUnit = useGameStore((s) => s.addUnit);
  const upgradeBuilding = useGameStore((s) => s.upgradeBuilding);
  const updateTick = useGameStore((s) => s.updateTick);

  const selectedUnitType = useUIStore((s) => s.selectedUnitType);
  const setSelectedUnitType = useUIStore((s) => s.setSelectedUnitType);

  const [notification, setNotification] = useState<string | null>(
    'Chào mừng đến Kingdom Arena! Nhấp thẻ quân bên dưới và chọn ô để xuất trận.'
  );

  // Initialize game on load
  useEffect(() => {
    initGame('player1');
  }, [initGame]);

  // Setup canvas & input
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const renderer = new GameRenderer(canvas);
    renderer.setMySide('player1');
    rendererRef.current = renderer;

    const input = new InputHandler(canvas, renderer);
    inputRef.current = input;

    input.setOnCellClick((col, row) => {
      const state = useGameStore.getState().gameState;
      const side = useGameStore.getState().mySide;
      const unitType = useUIStore.getState().selectedUnitType;

      if (!state || !side) return;

      const player = state.players[side];
      const validCols = side === 'player1' ? P1_SPAWN_COLS : P2_SPAWN_COLS;

      // 1. Check if clicking on an owned building to upgrade
      const clickedBuilding = player.buildings.find(
        (b) => b.position.col === col && b.position.row === row
      );

      if (clickedBuilding && !unitType) {
        if (clickedBuilding.level >= 3) {
          setNotification(`${clickedBuilding.type.toUpperCase()} đã đạt cấp tối đa (Lv3)!`);
          return;
        }

        const nextLevel = (clickedBuilding.level + 1) as 2 | 3;
        const bStats = BUILDING_STATS[clickedBuilding.type as 'castle' | 'fortress' | 'village'][nextLevel];
        const cost = (bStats as any)?.upgradeCost;

        if (
          cost &&
          player.resources.gold >= cost.gold &&
          player.resources.wood >= (cost.wood || 0) &&
          player.resources.stone >= (cost.stone || 0)
        ) {
          updateResources(side, {
            gold: player.resources.gold - cost.gold,
            wood: player.resources.wood - (cost.wood || 0),
            stone: player.resources.stone - (cost.stone || 0),
          });
          upgradeBuilding(clickedBuilding.id);
          setNotification(`🎉 Đã nâng cấp ${clickedBuilding.type.toUpperCase()} lên Lv${nextLevel}!`);
        } else {
          setNotification(
            `Không đủ tài nguyên nâng cấp! Cần: 🪙${cost?.gold || 0} 🪵${cost?.wood || 0} 🪨${cost?.stone || 0}`
          );
        }
        return;
      }

      // 2. Deploy unit if selected
      if (unitType) {
        if (!validCols.includes(col)) {
          setNotification('⚠️ Chỉ có thể thả quân trên nửa sân nhà của bạn!');
          return;
        }

        const stats = UNIT_STATS[unitType];
        if (
          player.resources.gold < stats.cost.gold ||
          player.resources.wood < (stats.cost.wood || 0) ||
          player.resources.stone < (stats.cost.stone || 0)
        ) {
          setNotification(`❌ Không đủ tài nguyên để xuất trận ${stats.name}!`);
          return;
        }

        // Deduct resources
        updateResources(side, {
          gold: player.resources.gold - stats.cost.gold,
          wood: player.resources.wood - (stats.cost.wood || 0),
          stone: player.resources.stone - (stats.cost.stone || 0),
        });

        // Spawn unit
        addUnit({
          id: `unit_${Date.now()}_${Math.random().toString(36).substring(2, 7)}`,
          type: unitType,
          owner: side,
          position: { col, row },
          hp: stats.hp,
          maxHp: stats.hp,
          atk: stats.atk,
          state: 'idle',
          lastAttackTime: 0,
        });

        setNotification(`⚔️ Đã xuất trận ${stats.name} tại ô (${col}, ${row})!`);
        setSelectedUnitType(null);
      }
    });

    return () => {
      input.destroy();
    };
  }, [addUnit, upgradeBuilding, updateResources, setSelectedUnitType]);

  // Sync deploy zone visual
  useEffect(() => {
    if (rendererRef.current) {
      rendererRef.current.setShowDeployZone(selectedUnitType !== null);
      if (mySide) rendererRef.current.setMySide(mySide);
    }
  }, [selectedUnitType, mySide]);

  // Core Game loop (Render 60fps + Resource Generation every 1s)
  useEffect(() => {
    if (!gameState) return;

    let lastResourceTick = Date.now();

    const loop = () => {
      const state = useGameStore.getState().gameState;
      if (!state || state.phase !== 'playing') return;

      const now = Date.now();
      if (now - lastResourceTick >= 1000) {
        lastResourceTick = now;

        for (const playerId of Object.keys(state.players)) {
          const player = state.players[playerId];
          let goldPerSec = 0,
            woodPerSec = 0,
            stonePerSec = 0;

          for (const b of player.buildings) {
            if (b.type === 'village' && b.hp > 0) {
              const prod = VILLAGE_PRODUCTION[b.level as 1 | 2 | 3];
              goldPerSec += prod.gold;
              woodPerSec += prod.wood;
              stonePerSec += prod.stone;
            }
          }

          updateResources(playerId, {
            gold: player.resources.gold + goldPerSec,
            wood: player.resources.wood + woodPerSec,
            stone: player.resources.stone + stonePerSec,
          });
        }

        updateTick(state.tick + 10, Math.max(0, state.timeRemaining - 1));
      }

      // Render board
      const currentState = useGameStore.getState().gameState;
      if (currentState && rendererRef.current) {
        rendererRef.current.render(currentState);
      }

      animFrameRef.current = requestAnimationFrame(loop);
    };

    animFrameRef.current = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(animFrameRef.current);
  }, [gameState?.phase, updateResources, updateTick]);

  return (
    <div
      style={{
        width: '100vw',
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'radial-gradient(ellipse at center, #1e1e38 0%, #0c0c16 100%)',
        position: 'relative',
        overflow: 'hidden',
        userSelect: 'none',
      }}
    >
      <ResourceBar />

      {notification && (
        <div
          style={{
            position: 'absolute',
            top: 60,
            padding: '6px 16px',
            background: 'rgba(30, 41, 59, 0.85)',
            border: '1px solid rgba(255, 255, 255, 0.15)',
            borderRadius: '16px',
            color: '#e2e8f0',
            fontSize: '12px',
            fontWeight: 500,
            zIndex: 15,
            pointerEvents: 'none',
            boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
          }}
        >
          {notification}
        </div>
      )}

      <div
        style={{
          boxShadow: '0 12px 48px rgba(0, 0, 0, 0.7)',
          borderRadius: '10px',
          overflow: 'hidden',
          border: '1px solid rgba(255, 255, 255, 0.1)',
        }}
      >
        <canvas
          ref={canvasRef}
          style={{
            display: 'block',
            maxWidth: '92vw',
            maxHeight: '75vh',
            cursor: selectedUnitType ? 'crosshair' : 'default',
          }}
        />
      </div>

      <UnitPanel />
    </div>
  );
}

export default App;
