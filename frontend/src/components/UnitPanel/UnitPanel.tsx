import { UnitType, UNIT_STATS } from '../../types/constants';
import { useUIStore } from '../../store/uiStore';
import { useGameStore } from '../../store/gameStore';

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

export function UnitPanel() {
  const selectedUnitType = useUIStore((s) => s.selectedUnitType);
  const setSelectedUnitType = useUIStore((s) => s.setSelectedUnitType);
  const setShowDeployZone = useUIStore((s) => s.setShowUpgradePanel);
  const gameState = useGameStore((s) => s.gameState);
  const mySide = useGameStore((s) => s.mySide);

  if (!gameState || !mySide) return null;

  const player = gameState.players[mySide];
  const resources = player?.resources;
  const units = Object.values(UnitType);

  // Highest fortress level determines unlocked units
  const maxFortressLevel = Math.max(
    1,
    ...player.buildings.filter(b => b.type === 'fortress').map(b => b.level)
  );

  return (
    <div style={{
      position: 'absolute',
      bottom: 16,
      left: '50%',
      transform: 'translateX(-50%)',
      display: 'flex',
      alignItems: 'center',
      gap: '10px',
      padding: '10px 18px',
      background: 'rgba(15, 23, 42, 0.9)',
      backdropFilter: 'blur(10px)',
      border: '1px solid rgba(255, 255, 255, 0.1)',
      borderRadius: '20px',
      boxShadow: '0 8px 32px rgba(0, 0, 0, 0.5)',
      zIndex: 20,
    }}>
      {units.map((type) => {
        const stats = UNIT_STATS[type];
        const isLocked = maxFortressLevel < stats.unlockLevel;
        const canAfford =
          !isLocked &&
          resources &&
          resources.gold >= stats.cost.gold &&
          resources.wood >= (stats.cost.wood || 0) &&
          resources.stone >= (stats.cost.stone || 0);

        const isSelected = selectedUnitType === type;

        return (
          <button
            key={type}
            disabled={isLocked}
            onClick={() => {
              const next = isSelected ? null : type;
              setSelectedUnitType(next);
              setShowDeployZone(next !== null);
            }}
            title={`${stats.name} (HP: ${stats.hp}, ATK: ${stats.atk})${isLocked ? ` - Yêu cầu Pháo Đài Lv${stats.unlockLevel}` : ''}`}
            style={{
              width: '68px',
              height: '84px',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'space-between',
              padding: '6px 4px',
              border: isSelected
                ? '2px solid #3b82f6'
                : canAfford
                ? '1px solid rgba(255, 255, 255, 0.2)'
                : '1px solid rgba(255, 255, 255, 0.05)',
              borderRadius: '12px',
              background: isSelected
                ? 'linear-gradient(180deg, rgba(59, 130, 246, 0.4), rgba(30, 64, 175, 0.6))'
                : isLocked
                ? 'rgba(0, 0, 0, 0.4)'
                : 'rgba(30, 41, 59, 0.7)',
              color: canAfford ? '#f8fafc' : '#64748b',
              cursor: isLocked ? 'not-allowed' : canAfford ? 'pointer' : 'default',
              opacity: isLocked ? 0.35 : canAfford ? 1 : 0.6,
              transform: isSelected ? 'translateY(-6px)' : 'none',
              transition: 'all 0.15s ease-out',
            }}
          >
            <div style={{ fontSize: '24px', lineHeight: 1 }}>{UNIT_ICONS[type]}</div>
            <div style={{ fontSize: '10px', fontWeight: 600, textAlign: 'center', whiteSpace: 'nowrap' }}>
              {stats.name}
            </div>
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: '2px',
              fontSize: '11px',
              fontWeight: 700,
              color: canAfford ? '#fbbf24' : '#94a3b8',
            }}>
              🪙 {stats.cost.gold}
            </div>
          </button>
        );
      })}
    </div>
  );
}
