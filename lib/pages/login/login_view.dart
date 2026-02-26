import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:get/get.dart';
import 'package:openim_common/openim_common.dart';

import 'login_logic.dart';

class LoginPage extends StatelessWidget {
  final logic = Get.find<LoginLogic>();

  LoginPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Material(
      child: TouchCloseSoftKeyboard(
        isGradientBg: true,
        child: SingleChildScrollView(
          child: Column(
            children: [
              88.verticalSpace,
              ImageRes.loginLogo.toImage
                ..width = 72.w
                ..height = 72.h
                ..onDoubleTap = logic.configService,
              12.verticalSpace,
              Text(
                '欢迎来到',
                style: TextStyle(
                  fontSize: 14.sp,
                  color: const Color(0xFF8E9AB0),
                  letterSpacing: 2,
                ),
              ),
              6.verticalSpace,
              Text(
                '南部战区办公室',
                style: TextStyle(
                  fontSize: 22.sp,
                  color: const Color(0xFF0C1C33),
                  fontWeight: FontWeight.w700,
                  letterSpacing: 3,
                ),
              ),
              8.verticalSpace,
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Container(
                    width: 30.w,
                    height: 1.h,
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        colors: [
                          Colors.transparent,
                          const Color(0xFF0089FF).withOpacity(0.4)
                        ],
                      ),
                    ),
                  ),
                  Container(
                    width: 4.w,
                    height: 4.w,
                    margin: EdgeInsets.symmetric(horizontal: 6.w),
                    decoration: const BoxDecoration(
                      color: Color(0xFF0089FF),
                      shape: BoxShape.circle,
                    ),
                  ),
                  Container(
                    width: 30.w,
                    height: 1.h,
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        colors: [
                          const Color(0xFF0089FF).withOpacity(0.4),
                          Colors.transparent
                        ],
                      ),
                    ),
                  ),
                ],
              ),
              40.verticalSpace,
              Padding(
                padding: EdgeInsets.symmetric(horizontal: 32.w),
                child: Obx(() => Column(
                      children: [
                        InputBox(
                          label: StrRes.username,
                          hintText: StrRes.plsEnterUsername,
                          controller: logic.usernameCtrl,
                          keyBoardType: TextInputType.text,
                        ),
                        16.verticalSpace,
                        InputBox.password(
                          label: StrRes.password,
                          hintText: StrRes.plsEnterPassword,
                          controller: logic.pwdCtrl,
                        ),
                        10.verticalSpace,
                        Row(
                          children: [
                            StrRes.forgetPassword.toText
                              ..style = Styles.ts_8E9AB0_12sp
                              ..onTap = logic.forgetPassword,
                          ],
                        ),
                        46.verticalSpace,
                        Button(
                          text: StrRes.login,
                          enabled: logic.enabled.value,
                          onTap: logic.login,
                        ),
                      ],
                    )),
              ),
              100.verticalSpace,
              RichText(
                text: TextSpan(
                  text: StrRes.noAccountYet,
                  style: Styles.ts_8E9AB0_12sp,
                  children: [
                    TextSpan(
                      text: StrRes.registerNow,
                      style: Styles.ts_0089FF_12sp,
                      recognizer: TapGestureRecognizer()
                        ..onTap = logic.registerNow,
                    )
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
