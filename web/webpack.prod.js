const config = require("./webpack.base.js");
const CssMinimizerPlugin = require("css-minimizer-webpack-plugin");
const UglifyJsPlugin = require("uglifyjs-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const CopyPlugin = require("copy-webpack-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const path = require("path");

config.mode = "production";
config.output.filename = () => "js/[name].[contenthash].js";

config.plugins.push(
  new MiniCssExtractPlugin({
    filename: "css/[name].[contenthash].css",
    chunkFilename: "css/[id].[chunkhash].css"
  })
);

config.optimization = {
  minimize: true,
  minimizer: [new UglifyJsPlugin(), new CssMinimizerPlugin()]
};

module.exports = config;
