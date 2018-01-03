import {connect} from 'react-redux';

import Component from 'components/Phases';

const mapStateToProps = (state) => {
  return {
    train: state.train.details,
    request: state.train.request
  };
};

export default connect(
  mapStateToProps
)(Component);
