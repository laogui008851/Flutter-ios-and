# 1. 安装 FVM (如未安装)
brew tap leoafarias/fvm && brew install fvm
# 2. 克隆项目后，在项目目录运行
fvm install
# 3. 使用项目锁定的 Flutter 版本fvm install
fvm flutter pub get
# 开发打包
fvm flutter run
# 安卓打包
fvm flutter build apk --release
# iOS 打包
fvm flutter build ios --release
fvm flutter build ipa --release
# windows 打包
fvm flutter build windows --release
# macOS 打包
fvm flutter build macos --release
# Linux 打包
fvm flutter build linux --release
# web 打包
fvm flutter build web --release
# 打包完成后，产物位于项目目录下的 build/ 目录中
# 4. 若要在命令行中全局使用 fvm 管理的 Flutter 版本，可将以下内容添加到 shell 配置文件中（如 .bashrc、.zshrc 等）
export PATH="$PATH":"$HOME/fvm/default/bin"
# 然后运行 source ~/.zshrc（或相应的配置文件）以使更改生效
source ~/.zshrc
# 5. 验证安装
fvm flutter --version
# 6. 若要切换 Flutter 版本，可使用以下命令
fvm use <version>
# 例如，切换到 Flutter 3.7.0 版本
fvm use 3.7.0 
# 7. 若要查看已安装的 Flutter 版本列表
fvm list
# 8. 若要卸载某个 Flutter 版本
fvm remove <version>
# 例如，卸载 Flutter 3.7.0 版本
fvm remove 3.7.0
# 9. 若要查看当前项目使用的 Flutter 版本
fvm flutter --version
# 10. 若要升级 FVM 到最新版本
brew upgrade fvm
# 11. 若要查看 FVM 帮助文档
fvm help
# 12. 若要查看项目中所有依赖包的最新版本
fvm flutter pub outdated
# 13. 若要升级项目中所有依赖包到最新版本
fvm flutter pub upgrade
# 14. 若要清理项目中的构建缓存
fvm flutter clean
# 15. 若要运行项目中的测试
fvm flutter test
# 16. 若要生成项目的发布版本
fvm flutter build <platform> --release
# 例如，生成安卓发布版本
fvm flutter build apk --release
# 17. 若要查看项目的依赖包树
fvm flutter pub deps
# 18. 若要查看项目的 Flutter 配置
fvm flutter config
# 19. 若要查看项目的 Flutter 环境信息
fvm flutter doctor
# 20. 若要查看项目的 Flutter 版本历史记录
fvm flutter version
# 21. 若要查看项目的 Flutter 渠道信息
fvm flutter channel
# 22. 若要切换项目的 Flutter 渠道
fvm flutter channel <channel>
# 例如，切换到 beta 渠道
fvm flutter channel beta
# 23. 若要查看项目的 Flutter 依赖包版本信息
fvm flutter pub version
# 24. 若要查看项目的 Flutter 依赖包许可证信息
fvm flutter pub licenses
# 25. 若要查看项目的 Flutter 依赖包缓存信息
fvm flutter pub cache
# 26. 若要查看项目的 Flutter 依赖包升级信息
fvm flutter pub upgrade --major-versions
# 27. 若要查看项目的 Flutter 依赖包修复信息
fvm flutter pub repair
# 28. 若要查看项目的 Flutter 依赖包依赖关系图
fvm flutter pub deps --graph
# 29. 若要查看项目的 Flutter 依赖包依赖关系树
fvm flutter pub deps --tree
# 30. 若要查看项目的 Flutter 依赖包依赖关系列表
fvm flutter pub deps --list