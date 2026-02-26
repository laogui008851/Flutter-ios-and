import 'package:flutter/material.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:get/get.dart';
import 'package:openim/widgets/register_page_bg.dart';
import 'package:openim_common/openim_common.dart';

import 'forget_password_logic.dart';

class ForgetPasswordPage extends StatelessWidget {
  final logic = Get.find<ForgetPasswordLogic>();

  ForgetPasswordPage({super.key});

  @override
  Widget build(BuildContext context) => RegisterBgView(
        child: Obx(() => Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                StrRes.forgetPassword.toText
                  ..style = Styles.ts_0089FF_22sp_semibold,
                29.verticalSpace,
                InputBox(
                  label: StrRes.username,
                  hintText: StrRes.plsEnterUsername,
                  controller: logic.usernameCtrl,
                  keyBoardType: TextInputType.text,
                ),
                16.verticalSpace,
                InputBox(
                  label: '授权码',
                  hintText: '请输入注册时的邀请码',
                  controller: logic.authCodeCtrl,
                ),
                130.verticalSpace,
                Button(
                  text: StrRes.nextStep,
                  enabled: logic.enabled.value,
                  onTap: logic.nextStep,
                ),
              ],
            )),
      );
}
