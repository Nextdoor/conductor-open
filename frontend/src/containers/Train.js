import {connect} from 'react-redux';

import Actions from 'actions';
import Component from 'components/Train';

const mapStateToProps = (state) => {
  return {
    train: state.train.details,
    request: state.train.request
  };
};

const mapDispatchToProps = (dispatch) => {
  return {
    load: (trainId) => {
      dispatch(Actions.Train.fetch(trainId));
    }
  };
};

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(Component);
