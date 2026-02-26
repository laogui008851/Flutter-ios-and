import 'package:flutter/material.dart';
import 'package:flutter_openim_sdk/flutter_openim_sdk.dart';
import 'package:get/get.dart';
import 'package:openim/pages/mine/server_config/server_config_binding.dart';
import 'package:openim/pages/mine/server_config/server_config_view.dart';
import 'package:openim_common/openim_common.dart';
import 'package:package_info_plus/package_info_plus.dart';

import '../../core/controller/im_controller.dart';
import '../../core/controller/push_controller.dart';
import '../../routes/app_navigator.dart';

// 保留 LoginType 枚举以兼容注册页等其他地方的引用
enum LoginType {
  phone,
  email,
}

extension LoginTypeExt on LoginType {
  int get rawValue => 0;

  String get name => StrRes.username;

  String get hintText => StrRes.plsEnterUsername;

  String get exclusiveName => StrRes.username;
}

class LoginLogic extends GetxController {
  final imLogic = Get.find<IMController>();
  final pushLogic = Get.find<PushController>();
  final usernameCtrl = TextEditingController();
  final pwdCtrl = TextEditingController();
  final enabled = false.obs;
  final versionInfo = ''.obs;
  LoginType operateType = LoginType.phone;

  String get username => usernameCtrl.text.trim();

  _initData() async {
    var map = DataSp.getLoginAccount();
    if (map is Map) {
      String? savedUsername = map["username"];
      if (savedUsername != null && savedUsername.isNotEmpty) {
        usernameCtrl.text = savedUsername;
      }
    }
  }

  @override
  void onClose() {
    usernameCtrl.dispose();
    pwdCtrl.dispose();
    super.onClose();
  }

  @override
  void onInit() {
    _initData();
    usernameCtrl.addListener(_onChanged);
    pwdCtrl.addListener(_onChanged);
    super.onInit();
  }

  @override
  void onReady() {
    super.onReady();
    getPackageInfo();
  }

  _onChanged() {
    enabled.value =
        usernameCtrl.text.trim().isNotEmpty && pwdCtrl.text.trim().isNotEmpty;
  }

  login() {
    LoadingView.singleton.wrap(asyncFunction: () async {
      var suc = await _login();
      if (suc) {
        Get.find<CacheController>().resetCache();
        AppNavigator.startMain();
      }
    });
  }

  Future<bool> _login() async {
    try {
      final uname = username;
      if (uname.isEmpty) {
        IMViews.showToast(StrRes.plsEnterUsername);
        return false;
      }
      if (!RegExp(r'^[a-zA-Z0-9]{3,20}$').hasMatch(uname)) {
        IMViews.showToast(StrRes.plsEnterRightUsername);
        return false;
      }
      final password = pwdCtrl.text;
      final data = await Apis.login(
        account: uname,
        password: password,
      );
      final account = {"username": uname};
      await DataSp.putLoginCertificate(data);
      await DataSp.putLoginAccount(account);
      Logger.print('login : ${data.userID}, token: ${data.imToken}');
      await imLogic.login(data.userID, data.imToken);
      Logger.print('im login success');
      pushLogic.login(data.userID);
      Logger.print('push login success');
      return true;
    } catch (e, s) {
      Logger.print('login e: $e $s');
    }
    return false;
  }

  void configService() => Get.to(
        () => ServerConfigPage(),
        binding: ServerConfigBinding(),
      );

  void registerNow() => AppNavigator.startRegister();

  void forgetPassword() => AppNavigator.startForgetPassword();

  void getPackageInfo() async {
    PackageInfo packageInfo = await PackageInfo.fromPlatform();
    final version = packageInfo.version;
    final appName = packageInfo.appName;
    final buildNumber = packageInfo.buildNumber;

    versionInfo.value = '$appName $version+$buildNumber SDK: ${OpenIM.version}';
  }
}
