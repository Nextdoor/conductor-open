import {connect} from 'react-redux';

import Component from 'components/Commits';

const mapStateToProps = (state) => {
  return {
    train: state.train.details,
    request: state.train.request
  };
};

export default connect(
  mapStateToProps
)(Component);
