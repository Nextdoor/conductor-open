import Actions from 'types/actions';
import API from 'api';

const request = (wipe) => {
  return {
    type: Actions.RequestTrain,
    wipe: wipe
  };
};

const receive = (train) => {
  return {
    type: Actions.ReceiveTrain,
    train: train,
    receivedAt: Date.now()
  };
};

const receiveError = (error) => {
  return {
    type: Actions.ReceiveTrainError,
    error: error,
    receivedAt: Date.now()
  };
};

const requestToggleClose = () => {
  return {
    type: Actions.RequestToggleClose
  };
};

const receiveToggleClose = (result) => {
  return {
    type: Actions.ReceiveToggleClose,
    closed: result.closed,
    scheduleOverride: result.schedule_override,
    receivedAt: Date.now()
  };
};

const receiveToggleCloseError = (error) => {
  return {
    type: Actions.ReceiveToggleCloseError,
    error: error,
    receivedAt: Date.now()
  };
};

const requestExtend = () => {
  return {
    type: Actions.RequestExtend
  };
};

const receiveExtend = () => {
  return {
    type: Actions.ReceiveExtend,
    receivedAt: Date.now()
  };
};

const receiveExtendError = (error) => {
  return {
    type: Actions.ReceiveExtendError,
    error: error,
    receivedAt: Date.now()
  };
};

const requestBlock = () => {
  return {
    type: Actions.RequestBlock
  };
};

const receiveBlock = () => {
  return {
    type: Actions.ReceiveBlock,
    receivedAt: Date.now()
  };
};

const receiveBlockError = (error) => {
  return {
    type: Actions.ReceiveBlockError,
    error: error,
    receivedAt: Date.now()
  };
};

const requestUnblock = () => {
  return {
    type: Actions.RequestUnblock
  };
};

const receiveUnblock = () => {
  return {
    type: Actions.ReceiveUnblock,
    receivedAt: Date.now()
  };
};

const receiveUnblockError = (error) => {
  return {
    type: Actions.ReceiveUnblockError,
    error: error,
    receivedAt: Date.now()
  };
};

const requestCancel = () => {
  return {
    type: Actions.RequestCancel
  };
};

const receiveCancel = () => {
  return {
    type: Actions.ReceiveCancel,
    receivedAt: Date.now()
  };
};

const receiveCancelError = (error) => {
  return {
    type: Actions.ReceiveCancelError,
    error: error,
    receivedAt: Date.now()
  };
};

const requestRollback = () => {
  return {
    type: Actions.RequestRollback
  };
};

const receiveRollback = () => {
  return {
    type: Actions.ReceiveRollback,
    receivedAt: Date.now()
  };
};

const receiveRollbackError = (error) => {
  return {
    type: Actions.ReceiveRollbackError,
    error: error,
    receivedAt: Date.now()
  };
};

const fetch = (trainId) => (dispatch) => {
  API.getTrain(trainId, dispatch, false);
};

const extend = (trainId) => (dispatch) => {
  API.extendTrain(trainId, dispatch);
};

const block = (trainId) => (dispatch) => {
  API.blockTrain(trainId, dispatch);
};

const unblock = (trainId) => (dispatch) => {
  API.unblockTrain(trainId, dispatch);
};

const cancel = (trainId) => (dispatch) => {
  API.cancelTrain(trainId, dispatch);
};

const rollbackTo = (trainId) => (dispatch) => {
  API.rollbackToTrain(trainId, dispatch);
};

const goToTrain = (trainId) => (dispatch) => {
  API.getTrain(trainId, dispatch, true);
};

const toggleClose = () => (dispatch, getState) => {
  const state = getState();
  const train = state.train.details;

  API.toggleClose(train.id, train.closed, dispatch);
};

export default {
  extend,
  block,
  unblock,
  cancel,
  rollbackTo,
  fetch,
  goToTrain,
  request,
  receive,
  receiveError,
  requestExtend,
  receiveExtend,
  receiveExtendError,
  requestBlock,
  receiveBlock,
  receiveBlockError,
  requestUnblock,
  receiveUnblock,
  receiveUnblockError,
  requestCancel,
  receiveCancel,
  receiveCancelError,
  requestRollback,
  receiveRollback,
  receiveRollbackError,
  requestToggleClose,
  receiveToggleClose,
  receiveToggleCloseError,
  toggleClose
};
