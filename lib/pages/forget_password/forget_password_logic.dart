import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:openim/routes/app_navigator.dart';
import 'package:openim_common/openim_common.dart';

class ForgetPasswordLogic extends GetxController {
  final usernameCtrl = TextEditingController();
  final authCodeCtrl = TextEditingController(); // 邀请码/授权码
  final enabled = false.obs;

  String get username => usernameCtrl.text.trim();

  @override
  void onClose() {
    usernameCtrl.dispose();
    authCodeCtrl.dispose();
    super.onClose();
  }

  @override
  void onInit() {
    usernameCtrl.addListener(_onChanged);
    authCodeCtrl.addListener(_onChanged);
    super.onInit();
  }

  _onChanged() {
    enabled.value = usernameCtrl.text.trim().isNotEmpty &&
        authCodeCtrl.text.trim().isNotEmpty;
  }

  /// 验证邀请码是否可用于找回密码
  Future<bool> verifyAuthCode() async {
    try {
      final result = await LoadingView.singleton.wrap(
        asyncFunction: () => Apis.verifyResetPasswordCode(
          code: authCodeCtrl.text.trim(),
          account: username,
        ),
      );
      return result == true;
    } catch (e) {
      return false;
    }
  }

  void nextStep() async {
    final uname = username;
    if (uname.isEmpty) {
      IMViews.showToast(StrRes.plsEnterUsername);
      return;
    }
    if (!RegExp(r'^[a-zA-Z0-9]{3,20}$').hasMatch(uname)) {
      IMViews.showToast(StrRes.plsEnterRightUsername);
      return;
    }

    final valid = await verifyAuthCode();
    if (!valid) {
      IMViews.showToast('授权码无效或不允许用于找回密码');
      return;
    }

    AppNavigator.startResetPassword(
      areaCode: '+00',
      phoneNumber: uname,
      email: null,
      verificationCode: authCodeCtrl.text.trim(),
    );
  }
}
