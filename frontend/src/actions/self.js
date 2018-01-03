import Actions from 'types/actions';
import API from 'api';

const request = () => {
  return {
    type: Actions.RequestSelf
  };
};

const receive = (self) => {
  return {
    type: Actions.ReceiveSelf,
    self: self,
    receivedAt: Date.now()
  };
};

const receiveError = (error) => {
  return {
    type: Actions.ReceiveSelfError,
    error: error,
    receivedAt: Date.now()
  };
};

const fetch = () => (dispatch) => {
  API.getSelf(dispatch);
};

export default {
  fetch,
  request,
  receive,
  receiveError
};
