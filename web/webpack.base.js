const path = require("path");
const CopyPlugin = require("copy-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");

module.exports = {
  devtool: "source-map",
  entry: {
    main: path.resolve(__dirname, "src/main.ts"),
    common: path.resolve(__dirname, "src/common.ts"),
    manage: path.resolve(__dirname, "src/Manage.tsx"),
    reader: path.resolve(__dirname, "src/Reader.tsx")
  },
  output: {
    clean: true,
    path: path.resolve("../bin/assets")
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        exclude: /node_modules/,
        use: ["babel-loader", "ts-loader"]
      },
      {
        test: /.less$/,
        use: [MiniCssExtractPlugin.loader, "css-loader", "postcss-loader", "less-loader"]
      },
      {
        test: /.css$/,
        use: [MiniCssExtractPlugin.loader, "css-loader", "postcss-loader"]
      }
    ]
  },
  plugins: [
    new CopyPlugin({
      patterns: [
        {
          from: path.resolve(__dirname, "src/templates"),
          to: path.resolve("../bin/templates")
        }
      ]
    }),
    new HtmlWebpackPlugin({
      filename: "../templates/head.html",
      template: path.resolve(__dirname, "src/head.html"),
      chunks: ["common", "main"],
      chunksSortMode: "manual",
      publicPath: "/"
    }),
    new HtmlWebpackPlugin({
      filename: "../templates/manage_head.html",
      template: path.resolve(__dirname, "src/manage_head.html"),
      chunks: ["common", "manage"],
      chunksSortMode: "manual",
      publicPath: "/"
    }),
    new HtmlWebpackPlugin({
      filename: "../templates/reader_head.html",
      template: path.resolve(__dirname, "src/reader_head.html"),
      chunks: ["common", "reader"],
      chunksSortMode: "manual",
      publicPath: "/"
    })
  ],
  resolve: {
    extensions: [".ts", ".tsx", ".js", ".jsx"]
  },
  watchOptions: {
    ignored: /node_modules/
  }
};
