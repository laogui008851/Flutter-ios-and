import 'package:flutter/material.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:get/get.dart';
import 'package:openim_common/openim_common.dart';

import 'splash_logic.dart';

class SplashPage extends StatelessWidget {
  final logic = Get.find<SplashLogic>();

  SplashPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            const Color(0xFFE8F4FF),
            const Color(0xFFFFFFFF),
          ],
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
        ),
      ),
      child: Stack(
        children: [
          // 顶部装饰条
          Positioned(
            top: 0,
            left: 0,
            right: 0,
            child: Container(
              height: 4.h,
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [
                    const Color(0xFF0055AA),
                    const Color(0xFF0089FF),
                    const Color(0xFF0055AA),
                  ],
                ),
              ),
            ),
          ),

          // 主体内容居中
          Center(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // 图标容器（加阴影）
                Container(
                  width: 110.w,
                  height: 110.w,
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(24.r),
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFF0089FF).withOpacity(0.18),
                        blurRadius: 24,
                        offset: const Offset(0, 8),
                      ),
                    ],
                  ),
                  child: ClipRRect(
                    borderRadius: BorderRadius.circular(24.r),
                    child: ImageRes.splashLogo.toImage
                      ..width = 110.w
                      ..height = 110.w,
                  ),
                ),

                28.verticalSpace,

                // 欢迎来到
                Text(
                  '欢迎来到',
                  style: TextStyle(
                    fontSize: 16.sp,
                    color: const Color(0xFF8E9AB0),
                    letterSpacing: 3,
                    fontWeight: FontWeight.w400,
                  ),
                ),

                8.verticalSpace,

                // 单位名称
                Text(
                  '南部战区办公室',
                  style: TextStyle(
                    fontSize: 26.sp,
                    color: const Color(0xFF0C1C33),
                    fontWeight: FontWeight.w700,
                    letterSpacing: 4,
                  ),
                ),

                16.verticalSpace,

                // 分隔线
                Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Container(
                      width: 40.w,
                      height: 1.h,
                      decoration: BoxDecoration(
                        gradient: LinearGradient(
                          colors: [
                            Colors.transparent,
                            const Color(0xFF0089FF).withOpacity(0.5),
                          ],
                        ),
                      ),
                    ),
                    Container(
                      width: 6.w,
                      height: 6.w,
                      margin: EdgeInsets.symmetric(horizontal: 8.w),
                      decoration: const BoxDecoration(
                        color: Color(0xFF0089FF),
                        shape: BoxShape.circle,
                      ),
                    ),
                    Container(
                      width: 40.w,
                      height: 1.h,
                      decoration: BoxDecoration(
                        gradient: LinearGradient(
                          colors: [
                            const Color(0xFF0089FF).withOpacity(0.5),
                            Colors.transparent,
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),

          // 底部版权
          Positioned(
            bottom: 40.h,
            left: 0,
            right: 0,
            child: Column(
              children: [
                Text(
                  '安全 · 高效 · 保密',
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    fontSize: 12.sp,
                    color: const Color(0xFF8E9AB0),
                    letterSpacing: 2,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
