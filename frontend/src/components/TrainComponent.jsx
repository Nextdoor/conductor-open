import React from 'react';

import {trainProps, requestProps} from 'types/proptypes';
import Error from 'components/Error';
import Loading from 'components/Loading';

class TrainComponent extends React.Component {
  getRequestComponent() {
    const {request, train} = this.props;
    if (train === null && (request.fetching === true || request.receivedAt === null)) {
      return <Loading/>;
    }

    if (request.error !== null) {
      return <Error message={request.error}/>;
    }

    return null;
  }
}

TrainComponent.propTypes = {
  train: trainProps,
  request: requestProps
};

export default TrainComponent;
