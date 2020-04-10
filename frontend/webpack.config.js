const webpack = require('webpack');
const path = require('path');

const dir = path.join(__dirname, '..', 'resources', 'frontend', 'gen');

const IS_PRODUCTION = process.env.IS_PRODUCTION === 'true';

const REQUIRED_ENV_VARS = [
  'OAUTH_PROVIDER',
  'OAUTH_ENDPOINT',

  // json payload for the oauth request, e.g.
  // {"client_id": "123",
  //  "redirect_uri": "<hostname>/api/auth/done",
  //  "scope": "commit"}
  'OAUTH_PAYLOAD'
];
const DEFAULT_ENV_VARS = {
  'AUTH_COOKIE_NAME': 'conductor-auth'
};

// Look for missing env vars.
let missing_env = false;
REQUIRED_ENV_VARS.forEach(function (env_var) {
  if (!(env_var in process.env)) {
    console.error(
      'Environmental variable ' + env_var +
      ' must be set for Webpack build.');
    missing_env = true;
  } else {
    console.log(env_var + ':');
    console.log(process.env[env_var]);
  }
});
if (missing_env) {
  process.exit(1);
}

for (const env_var in DEFAULT_ENV_VARS) {
  if (!(env_var in process.env)) {
    // Set to default value.
    process.env[env_var] = DEFAULT_ENV_VARS[env_var];
  }
  console.log(env_var + ':');
  console.log(process.env[env_var]);
}

const env_vars = REQUIRED_ENV_VARS.concat(Object.keys(DEFAULT_ENV_VARS));

const PLUGIN_CONFIG = [
  new webpack.EnvironmentPlugin(env_vars)
];

let DEV_TOOL = 'source-map';

if (IS_PRODUCTION) {
  DEV_TOOL = 'false';
}

/** Webpack Config */
module.exports = {
  entry: ['whatwg-fetch', './src/index.jsx'],
  output: {
    path: dir,
    filename: 'bundle.js'
  },
  resolve: {
    extensions: ['.js', '.jsx'],
    modules: [
      path.join(__dirname, "src"),
      "node_modules"
    ]
  },
  module: {
    rules: [
      {
        test: /\.jsx$/,
        loader: "babel-loader", // Do not use "use" here
        exclude: /(node_modules)/,
        query: {
          presets: ['@babel/react', '@babel/env']
        }
      },
      {
        test: /\.css$/,
        use: [
          {
            loader: "style-loader"
          },
          {
            loader: "css-loader",
          },
          {
            loader: "sass-loader"
          }
        ]
      },
      {
        test: /\.scss$/,
        use: [
          {
            loader: "style-loader"
          },
          {
            loader: "css-loader"
          },
          {
            loader: "sass-loader"
          }
        ]
      },
      {
        test: /\.(woff(2)?|ttf|eot|svg)(\?v=\d+\.\d+\.\d+)?$/,
        use: [
          {
            loader: 'url-loader',
            options: {
              limit: '10000',
              name: '/[hash].[ext]',
            }
          }
        ]
      }
    ]
  },
  plugins: PLUGIN_CONFIG,
  devtool: DEV_TOOL,
  performance: {
    hints: false
  }
};
