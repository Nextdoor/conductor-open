import moment from 'moment';

import Actions from 'types/actions';

const train = (state = null, action) => {
  let newState = state;
  switch (action.type) {
    case Actions.RequestTrain:
      newState = Object.assign({}, state, {
        request: {
          fetching: true,
          error: null,
          receivedAt: null
        }
      });
      if (action.wipe) {
        newState = Object.assign({}, newState, {
          details: null,
        });
      }
      break;
    case Actions.ReceiveTrain:
      newState = Object.assign({}, state, {
        details: action.train,
        request: {
          fetching: false,
          error: null,
          receivedAt: action.receivedAt
        }
      });
      break;
    case Actions.ReceiveTrainError:
      newState = Object.assign({}, state, {
        details: null,
        request: {
          fetching: false,
          error: action.error,
          receivedAt: action.receivedAt
        }
      });
      break;
    case Actions.RequestToggleClose:
      newState = Object.assign({}, state, {
        requestToggleClose: {
          fetching: true,
          error: null,
          receivedAt: null
        }
      });
      break;
    case Actions.ReceiveToggleClose:
      newState = Object.assign({}, state, {
        details: Object.assign({}, state.details, {
          closed: action.closed,
          schedule_override: action.scheduleOverride
        }),
        requestToggleClose: {
          fetching: false,
          error: null,
          receivedAt: action.receivedAt
        }
      });
      break;
    case Actions.ReceiveToggleCloseError:
      newState = Object.assign({}, state, {
        requestToggleClose: {
          fetching: false,
          error: action.error,
          receivedAt: action.receivedAt
        }
      });
      break;
    case Actions.RequestExtend:
      newState = Object.assign({}, state, {
        requestExtend: {
          fetching: true,
          error: null,
          receivedAt: null
        }
      });
      break;
    case Actions.ReceiveExtend:
      newState = Object.assign({}, state, {
        requestExtend: {
          fetching: false,
          error: null,
          receivedAt: action.receivedAt
        }
      });
      break;
    case Actions.ReceiveExtendError:
      newState = Object.assign({}, state, {
        requestExtend: {
          fetching: false,
          error: action.error,
          receivedAt: action.receivedAt
        }
      });
      break;
    case Actions.RequestBlock:
      newState = Object.assign({}, state, {
        requestBlock: {
          fetching: true,
          error: null,
          receivedAt: null
        }
      });
      break;
    case Actions.ReceiveBlock:
      newState = Object.assign({}, state, {
        details: Object.assign({}, state.details, {
          blocked: true
        }),
        requestBlock: {
          fetching: false,
          error: null,
          receivedAt: action.receivedAt
        }
      });
      break;
    case Actions.ReceiveBlockError:
      newState = Object.assign({}, state, {
        requestBlock: {
          fetching: false,
          error: action.error,
          receivedAt: action.receivedAt
        }
      });
      break;
    case Actions.RequestUnblock:
      newState = Object.assign({}, state, {
        requestUnblock: {
          fetching: true,
          error: null,
          receivedAt: null
        }
      });
      break;
    case Actions.ReceiveUnblock:
      newState = Object.assign({}, state, {
        details: Object.assign({}, state.details, {
          blocked: false
        }),
        requestUnblock: {
          fetching: false,
          error: null,
          receivedAt: action.receivedAt
        }
      });
      break;
    case Actions.ReceiveUnblockError:
      newState = Object.assign({}, state, {
        requestUnblock: {
          fetching: false,
          error: action.error,
          receivedAt: action.receivedAt
        }
      });
      break;
    case Actions.RequestCancel:
      newState = Object.assign({}, state, {
        requestCancel: {
          fetching: true,
          error: null,
          receivedAt: null
        }
      });
      break;
    case Actions.ReceiveCancel:
      newState = Object.assign({}, state, {
        details: Object.assign({}, state.details, {
          cancelled_at: moment().format(),
          done: true
        }),
        requestCancel: {
          fetching: false,
          error: null,
          receivedAt: action.receivedAt
        }
      });
      break;
    case Actions.ReceiveCancelError:
      newState = Object.assign({}, state, {
        requestCancel: {
          fetching: false,
          error: action.error,
          receivedAt: action.receivedAt
        }
      });
      break;
    default:
      break;
  }
  return newState;
};

export default train;
