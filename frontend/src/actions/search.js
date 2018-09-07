import Actions from 'types/actions';
import API from 'api';

const request = () => {
  return {
    type: Actions.RequestSearch
  };
};

const receive = (search) => {
  return {
    type: Actions.ReceiveSearch,
    search: search,
    receivedAt: Date.now(),
    searchQuery: search.params.commit
  };
};

const receiveError = (error) => {
  return {
    type: Actions.ReceiveSearchError,
    error: error,
    receivedAt: Date.now()
  };
};

const fetch = (params) => (dispatch) => {
  API.getSearch(dispatch, params);
};

export default {
  fetch,
  request,
  receive,
  receiveError
};
