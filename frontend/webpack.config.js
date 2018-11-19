const webpack = require('webpack');
const path = require('path');
const ParallelUglifyPlugin = require('webpack-parallel-uglify-plugin');

const dir = path.join(__dirname, '..', 'resources', 'frontend', 'gen');

const IS_PRODUCTION = process.env.IS_PRODUCTION === 'true';

const PLUGIN_CONFIG = [];

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
