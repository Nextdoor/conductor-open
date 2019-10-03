import {connect} from 'react-redux';

import Actions from 'actions';
import Component from 'components/Details';

const mapStateToProps = (state) => {
  return {
    train: state.train.details,
    request: state.train.request,
    requestExtend: state.train.requestExtend,
    requestBlock: state.train.requestBlock,
    requestUnblock: state.train.requestUnblock
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
    changeEngineer: (trainId) => {
      dispatch(Actions.Train.changeEngineer(trainId));
    }
  };
};

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(Component);

