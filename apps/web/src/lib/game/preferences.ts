export interface GamePreferences {
  inputMode: 'cell-first' | 'number-first';
  soundEnabled: boolean;
  hapticsEnabled: boolean;
}

export const defaultGamePreferences: GamePreferences = {
  inputMode: 'cell-first',
  soundEnabled: false,
  hapticsEnabled: false,
};

export function loadGamePreferences(storage: Pick<Storage, 'getItem'>): GamePreferences {
  return {
    ...defaultGamePreferences,
    inputMode:
      storage.getItem('ninefold.inputMode') === 'number-first' ? 'number-first' : 'cell-first',
    soundEnabled: storage.getItem('ninefold.sound') === 'on',
    hapticsEnabled: storage.getItem('ninefold.haptics') === 'on',
  };
}
