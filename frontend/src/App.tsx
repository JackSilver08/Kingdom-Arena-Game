import { useEffect, useRef, useState } from 'react';
import { GameRenderer } from './game/renderer';
import { InputHandler } from './game/input';
import { useGameStore } from './store/gameStore';
import { useUIStore } from './store/uiStore';
import { ResourceBar } from './components/HUD/ResourceBar';
import { UnitPanel } from './components/UnitPanel/UnitPanel';
import { WebSocketService } from './services/websocket';

function createCommandId() {
  return `cmd_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`;
}

function App() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const rendererRef = useRef<GameRenderer | null>(null);
  const socketRef = useRef<WebSocketService | null>(null);

  const gameState = useGameStore((state) => state.gameState);
  const mySide = useGameStore((state) => state.mySide);
  const setGameState = useGameStore((state) => state.setGameState);
  const setMySide = useGameStore((state) => state.setMySide);

  const selectedUnitType = useUIStore((state) => state.selectedUnitType);
  const selectedUnitId = useUIStore((state) => state.selectedUnitId);
  const commandTarget = useUIStore((state) => state.commandTarget);
  const setSelectedUnitType = useUIStore((state) => state.setSelectedUnitType);
  const setSelectedUnitId = useUIStore((state) => state.setSelectedUnitId);
  const setCommandTarget = useUIStore((state) => state.setCommandTarget);

  const [notification, setNotification] = useState('Đang kết nối máy chủ trận đấu...');

  // The Go engine owns the game state. The browser sends commands and renders snapshots.
  useEffect(() => {
    const socket = new WebSocketService();
    socketRef.current = socket;

    socket.onMessage((message) => {
      switch (message.type) {
        case 'GAME_START':
          setGameState(message.payload.gameState);
          setMySide(message.payload.yourSide);
          setNotification(
            `Đã vào trận với phe ${message.payload.yourSide === 'player1' ? 'Xanh' : 'Đỏ'}. Chuột trái chọn quân, chuột phải ra lệnh.`
          );
          break;
        case 'STATE_UPDATE':
          setGameState(message.payload.gameState);
          break;
        case 'COMMAND_REJECTED':
          setNotification(`Lệnh bị từ chối: ${message.payload.message}`);
          break;
        case 'PONG':
          break;
      }
    });

    socket.connect().catch(() => {
      setNotification('Chưa kết nối được máy chủ. Đang thử kết nối lại...');
    });

    return () => {
      socket.disconnect();
      socketRef.current = null;
    };
  }, [setGameState, setMySide]);

  // Set up left-click selection, right-click orders and deployment.
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const renderer = new GameRenderer(canvas);
    rendererRef.current = renderer;
    const input = new InputHandler(canvas, renderer);

    input.setOnCellClick((col, row) => {
      const state = useGameStore.getState().gameState;
      const side = useGameStore.getState().mySide;
      const unitType = useUIStore.getState().selectedUnitType;

      if (!state || !side) {
        setNotification('Đang chờ đồng bộ trận đấu với máy chủ...');
        return;
      }

      const ownPlayer = state.players[side];
      const ownUnit = ownPlayer.units.find(
        (unit) => unit.position.col === col && unit.position.row === row && unit.hp > 0
      );
      if (ownUnit) {
        setSelectedUnitId(ownUnit.id);
        setCommandTarget(null);
        setNotification(
          `Đã chọn ${ownUnit.type}. Nhấn chuột phải vào ô muốn di chuyển hoặc tấn công.`
        );
        return;
      }

      const enemyUnit = Object.entries(state.players)
        .filter(([playerId]) => playerId !== side)
        .flatMap(([, player]) => player.units)
        .find((unit) => unit.position.col === col && unit.position.row === row && unit.hp > 0);
      if (enemyUnit) {
        setNotification('Đây là quân đối phương. Hãy chọn quân phe mình rồi nhấn chuột phải để tấn công.');
        return;
      }

      const ownBuilding = ownPlayer.buildings.find(
        (building) =>
          building.position.col === col &&
          building.position.row === row &&
          building.hp > 0
      );
      if (ownBuilding) {
        if (unitType) {
          setNotification('Không thể triển khai quân trên ô đang có công trình.');
          return;
        }
        const sent = socketRef.current?.send({
          type: 'UPGRADE',
          payload: { buildingId: ownBuilding.id, commandId: createCommandId() },
        });
        setNotification(
          sent
            ? `Đang gửi lệnh nâng cấp ${ownBuilding.type}...`
            : 'Mất kết nối máy chủ. Chưa gửi được lệnh.'
        );
        return;
      }

      const enemyBuilding = Object.entries(state.players)
        .filter(([playerId]) => playerId !== side)
        .flatMap(([, player]) => player.buildings)
        .find(
          (building) =>
            building.position.col === col &&
            building.position.row === row &&
            building.hp > 0
        );
      if (enemyBuilding) {
        setNotification('Muốn đánh công trình? Chọn một quân rồi nhấn chuột phải vào công trình đó.');
        return;
      }

      if (unitType) {
        const sent = socketRef.current?.send({
          type: 'DEPLOY_UNIT',
          payload: { unitType, col, row, commandId: createCommandId() },
        });
        if (sent) {
          setNotification(`Đang yêu cầu triển khai ${unitType} tại ô (${col}, ${row})...`);
          setSelectedUnitType(null);
        } else {
          setNotification('Mất kết nối máy chủ. Chưa gửi được lệnh triển khai.');
        }
        return;
      }

      setSelectedUnitId(null);
      setCommandTarget(null);
    });

    input.setOnCellRightClick((col, row) => {
      const state = useGameStore.getState().gameState;
      const side = useGameStore.getState().mySide;
      const unitId = useUIStore.getState().selectedUnitId;
      if (!state || !side) {
        setNotification('Đang chờ kết nối máy chủ...');
        return;
      }
      if (!unitId) {
        setNotification('Hãy nhấn chuột trái chọn một quân trước.');
        return;
      }

      const ownPlayer = state.players[side];
      const selectedUnit = ownPlayer.units.find((unit) => unit.id === unitId && unit.hp > 0);
      if (!selectedUnit) {
        setSelectedUnitId(null);
        setNotification('Quân đã chọn không còn tồn tại. Hãy chọn lại.');
        return;
      }

      const friendlyUnit = ownPlayer.units.some(
        (unit) => unit.position.col === col && unit.position.row === row && unit.hp > 0
      );
      const friendlyBuilding = ownPlayer.buildings.some(
        (building) =>
          building.position.col === col &&
          building.position.row === row &&
          building.hp > 0
      );
      if (friendlyUnit || friendlyBuilding) {
        setNotification('Không thể ra lệnh tấn công quân hoặc công trình của phe mình.');
        return;
      }

      const sent = socketRef.current?.send({
        type: 'MOVE_TO',
        payload: { unitId, col, row, commandId: createCommandId() },
      });
      if (sent) {
        setCommandTarget({ col, row });
        setNotification(
          `Đã gửi lệnh cho ${selectedUnit.type}: di chuyển đến (${col}, ${row}) hoặc tấn công mục tiêu tại đó.`
        );
      } else {
        setNotification('Mất kết nối máy chủ. Chưa gửi được lệnh điều quân.');
      }
    });

    return () => {
      input.destroy();
      rendererRef.current = null;
    };
  }, [setSelectedUnitId, setSelectedUnitType, setCommandTarget]);

  // Reflect selection and last order visually on the canvas.
  useEffect(() => {
    const renderer = rendererRef.current;
    if (!renderer) return;
    renderer.setMySide(mySide || 'player1');
    renderer.setShowDeployZone(selectedUnitType !== null);
    renderer.setSelectedUnitId(selectedUnitId);
    renderer.setCommandTarget(commandTarget);
  }, [mySide, selectedUnitType, selectedUnitId, commandTarget]);

  // Render continuously; the authoritative state arrives from the server.
  useEffect(() => {
    let frame = 0;
    const render = () => {
      const state = useGameStore.getState().gameState;
      if (state && rendererRef.current) rendererRef.current.render(state);
      frame = requestAnimationFrame(render);
    };
    frame = requestAnimationFrame(render);
    return () => cancelAnimationFrame(frame);
  }, []);

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
            background: 'rgba(30, 41, 59, 0.9)',
            border: '1px solid rgba(255, 255, 255, 0.15)',
            borderRadius: '16px',
            color: '#e2e8f0',
            fontSize: '12px',
            fontWeight: 500,
            zIndex: 15,
            pointerEvents: 'none',
            boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
            textAlign: 'center',
            maxWidth: '90vw',
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
            cursor: selectedUnitType || selectedUnitId ? 'crosshair' : 'default',
          }}
        />
      </div>

      {gameState && <UnitPanel />}
    </div>
  );
}

export default App;
