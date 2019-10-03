import React from 'react';
import {SyncIcon} from 'react-octicons';
import ReactCSSTransitionGroup from 'react-addons-css-transition-group';

const Loading = () => {
  return (
    <span className="loading">
      <ReactCSSTransitionGroup
        transitionName="loading"
        transitionAppear
        transitionEnterTimeout={20}
        transitionAppearTimeout={20}
        transitionLeaveTimeout={20}>
        <SyncIcon className="spin-octicon" key="animation"/>
      </ReactCSSTransitionGroup>
    </span>
  );
};

export default Loading;
