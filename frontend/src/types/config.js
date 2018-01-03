export const Modes = {
  Schedule: 0,
  Manual: 1,
};

export function modeToString(mode) {
  switch (mode) {
    case Modes.Schedule:
      return 'schedule';
    case Modes.Manual:
      return 'manual';
    default:
      return null;
  }
}

export function stringToMode(modeString) {
  switch (modeString) {
    case 'schedule':
      return Modes.Schedule;
    case 'manual':
      return Modes.Manual;
    default:
      return null;
  }
}
