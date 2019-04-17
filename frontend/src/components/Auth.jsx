import React from 'react';
import {MarkGithubIcon} from 'react-octicons';

import API from 'api';
import Card from './Card';

const oauthProvider = process.env.OAUTH_PROVIDER;
const oauthEndpoint = process.env.OAUTH_ENDPOINT;
const oauthPayload = JSON.parse(process.env.OAUTH_PAYLOAD);

const oauthURL = oauthEndpoint + '?' + API.encodeQueryParams(oauthPayload);

const Auth = () => {
  let icon;
  if (oauthProvider.toLowerCase() === "github") {
    icon = <MarkGithubIcon className="login-provider-icon"/>;
  }
  const header = (
    <div className="login-header-text">
      <p>Welcome to Conductor</p>
      <p>All aboard!</p>
    </div>
  );
  return (
    <Card header={header} className="login-card">
      <a href={oauthURL}>
        <button className="login-button">
          {icon}
          <div className="login-button-text">
            Login with {oauthProvider}
          </div>
        </button>
      </a>
    </Card>
  );
};

export default Auth;
