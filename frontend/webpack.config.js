const webpack = require('webpack');
const path = require('path');
const ParallelUglifyPlugin = require('webpack-parallel-uglify-plugin');

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
REQUIRED_ENV_VARS.forEach(function(env_var) {
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
  PLUGIN_CONFIG.push(new ParallelUglifyPlugin({uglifyJS: {minimize: true}}));
  PLUGIN_CONFIG.push(new webpack.DefinePlugin({
    'process.env': {
      'NODE_ENV': '"production"'
    }
  }));

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
    extensions: ['', '.js', '.jsx'],
    root: [path.resolve('./src')]
  },
  module: {
    loaders: [
      {
        test: /\.jsx?$/,
        loader: 'babel-loader',
        exclude: /(node_modules)/,
        query: {
          presets: ['es2015', 'react']
        },
        progress: true
      },
      {test: /\.css$/, loader: 'style!css'},
      {test: /\.scss$/, loaders: ['style-loader', 'css-loader', 'sass-loader']},
      {test: /\.(otf|eot|svg|ttf|woff|woff2).*$/, loader: 'url?limit=8192&name=/[hash].[ext]'}
    ]
  },
  plugins: PLUGIN_CONFIG,
  devtool: DEV_TOOL
};
