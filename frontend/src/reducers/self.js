import Actions from 'types/actions';

const self = (state = null, action) => {
  switch (action.type) {
    case Actions.RequestSelf:
      return Object.assign({}, state, {
        details: null,
        request: {
          fetching: true,
          error: null,
          receivedAt: null
        }
      });
    case Actions.ReceiveSelf:
      return Object.assign({}, state, {
        details: action.self,
        request: {
          fetching: false,
          error: null,
          receivedAt: action.receivedAt
        }
      });
    case Actions.ReceiveSelfError:
      return Object.assign({}, state, {
        details: null,
        request: {
          fetching: false,
          error: action.error,
          receivedAt: action.receivedAt
        }
      });
    default:
      return state;
  }
};

export default self;
