const config = require("./webpack.base.js");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const CopyPlugin = require("copy-webpack-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const path = require("path");

config.mode = "development";
config.output.filename = () => "js/[name].development.js";

config.plugins.push(
  new MiniCssExtractPlugin({
    filename: "css/[name].development.css",
    chunkFilename: "css/[id].development.css"
  })
);

module.exports = config;
