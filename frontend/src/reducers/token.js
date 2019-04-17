import Actions from 'types/actions';

const token = (state = null, action) => {
  switch (action.type) {
    case Actions.SetToken:
      return Object.assign({}, state, {
        token: action.token,
        promptLogin: false
      });
    case Actions.PromptLogin:
      return Object.assign({}, state, {
        promptLogin: true
      });
    case Actions.DeleteToken:
      return Object.assign({}, state, {
        token: null
      });
    case Actions.RequestLogout:
      return Object.assign({}, state, {
        logoutRequest: {
          fetching: true,
          error: null,
          receivedAt: null
        }
      });
    case Actions.ReceiveLogout:
      return Object.assign({}, state, {
        logoutRequest: {
          fetching: false,
          error: null,
          receivedAt: action.receivedAt
        }
      });
    case Actions.ReceiveLogoutError:
      return Object.assign({}, state, {
        logoutRequest: {
          fetching: false,
          error: action.error,
          receivedAt: action.receivedAt
        }
      });
    default:
      return state;
  }
};

export default token;
