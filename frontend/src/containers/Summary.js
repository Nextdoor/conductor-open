import {connect} from 'react-redux';

import Actions from 'actions';
import Component from 'components/Summary';

const mapStateToProps = (state) => {
  return {
    train: state.train.details,
    request: state.train.request,
    requestExtend: state.train.requestExtend,
    requestBlock: state.train.requestBlock,
    requestUnblock: state.train.requestUnblock,
    requestCancel: state.train.requestCancel,
    requestRollback: state.train.requestRollback
  };
};

const mapDispatchToProps = (dispatch) => {
  return {
    extendTrain: (trainId) => {
      dispatch(Actions.Train.extend(trainId));
    },
    blockTrain: (trainId) => {
      dispatch(Actions.Train.block(trainId));
    },
    unblockTrain: (trainId) => {
      dispatch(Actions.Train.unblock(trainId));
    },
    cancelTrain: (trainId) => {
      dispatch(Actions.Train.cancel(trainId));
    },
    rollbackToTrain: (trainId) => {
      dispatch(Actions.Train.rollbackTo(trainId));
    }
  };
};

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(Component);
