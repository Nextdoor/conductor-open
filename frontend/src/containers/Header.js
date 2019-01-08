import {connect} from 'react-redux';

import Actions from 'actions';
import Component from 'components/Header';

const mapStateToProps = (state) => {
  return {
    self: state.self.details,
    request: state.self.request,
    train: state.train.details
  };
};

const mapDispatchToProps = (dispatch) => {
  return {
    load: () => {
      dispatch(Actions.Self.fetch());
    },
  };
};

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(Component);
