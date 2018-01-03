import Actions from 'types/actions';

const search = (state = null, action) => {
  let newState = state;
  switch (action.type) {
    case Actions.RequestSearch:
      newState = Object.assign({}, state, {
        request: {
          fetching: true,
          error: null,
          receivedAt: null
        }
      });
      break;
    case Actions.ReceiveSearch:
      newState = Object.assign({}, state, {
        details: action.search,
        request: {
          fetching: false,
          error: null,
          receivedAt: action.receivedAt
        }
      });
      break;
    case Actions.ReceiveSearchError:
      newState = Object.assign({}, state, {
        details: null,
        request: {
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

export default search;
