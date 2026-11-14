import { useGameStore } from '../../store/gameStore';

export function ResourceBar() {
  const gameState = useGameStore((s) => s.gameState);
  const mySide = useGameStore((s) => s.mySide);

  if (!gameState || !mySide) return null;

  const player = gameState.players[mySide];
  const resources = player?.resources;
  if (!resources) return null;

  const minutes = Math.floor(gameState.timeRemaining / 60);
  const seconds = gameState.timeRemaining % 60;

  // Count active villages
  const activeVillages = player.buildings.filter(b => b.type === 'village' && b.hp > 0).length;

  return (
    <div style={{
      position: 'absolute',
      top: 12,
      left: '50%',
      transform: 'translateX(-50%)',
      display: 'flex',
      alignItems: 'center',
      gap: '20px',
      padding: '8px 24px',
      background: 'rgba(15, 23, 42, 0.85)',
      backdropFilter: 'blur(8px)',
      border: '1px solid rgba(255, 255, 255, 0.1)',
      borderRadius: '24px',
      color: '#f8fafc',
      fontFamily: 'system-ui, -apple-system, sans-serif',
      fontSize: '14px',
      fontWeight: 600,
      zIndex: 20,
      boxShadow: '0 4px 20px rgba(0, 0, 0, 0.4)',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
        <span style={{ fontSize: '18px' }}>🪙</span>
        <span>{Math.floor(resources.gold)}</span>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
        <span style={{ fontSize: '18px' }}>🪵</span>
        <span>{Math.floor(resources.wood)}</span>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
        <span style={{ fontSize: '18px' }}>🪨</span>
        <span>{Math.floor(resources.stone)}</span>
      </div>

      <div style={{
        width: '1px',
        height: '20px',
        background: 'rgba(255, 255, 255, 0.15)',
      }} />

      <div style={{ display: 'flex', alignItems: 'center', gap: '6px', color: '#94a3b8', fontSize: '12px' }}>
        <span>🏘️ {activeVillages}/3 Làng</span>
      </div>

      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: '6px',
        background: 'rgba(239, 68, 68, 0.2)',
        padding: '3px 10px',
        borderRadius: '12px',
        border: '1px solid rgba(239, 68, 68, 0.4)',
        color: '#fca5a5',
      }}>
        <span>⏱️</span>
        <span>{minutes}:{seconds.toString().padStart(2, '0')}</span>
      </div>
    </div>
  );
}
