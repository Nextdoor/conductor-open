import Actions from 'types/actions';

const admin = (state = null, action) => {
  switch (action.type) {
    case Actions.RequestConfig:
      return Object.assign({}, state, {
        requestConfig: {
          fetching: true,
          error: null,
          receivedAt: null
        }
      });
    case Actions.ReceiveConfig:
      return Object.assign({}, state, {
        config: action.config,
        requestConfig: {
          fetching: false,
          error: null,
          receivedAt: action.receivedAt
        }
      });
    case Actions.ReceiveConfigError:
      return Object.assign({}, state, {
        config: null,
        requestConfig: {
          fetching: false,
          error: action.error,
          receivedAt: action.receivedAt
        }
      });
    case Actions.RequestToggleMode:
      return Object.assign({}, state, {
        requestToggleMode: {
          fetching: true,
          error: null,
          receivedAt: null
        }
      });
    case Actions.ReceiveToggleMode:
      return Object.assign({}, state, {
        config: Object.assign({}, state.config, {
          mode: action.mode
        }),
        requestToggleMode: {
          fetching: false,
          error: null,
          receivedAt: action.receivedAt
        }
      });
    case Actions.ReceiveToggleModeError:
      return Object.assign({}, state, {
        requestToggleMode: {
          fetching: false,
          error: action.error,
          receivedAt: action.receivedAt
        }
      });
    default:
      return state;
  }
};

export default admin;
