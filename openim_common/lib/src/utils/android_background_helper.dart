import 'dart:io';

import 'package:device_info_plus/device_info_plus.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:get/get.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../res/strings.dart';
import '../res/styles.dart';

/// Android 后台运行权限引导工具
/// 帮助用户在国产手机上开启自启动权限和关闭电池优化
class AndroidBackgroundHelper {
  AndroidBackgroundHelper._();

  static const String _prefKeyShownAutoStart = 'shown_auto_start_guide';
  static const String _prefKeyShownBattery = 'shown_battery_guide';

  /// 厂商名称到包名的映射
  static const Map<String, String> _manufacturerIntents = {
    'xiaomi': 'com.miui.securitycenter',
    'redmi': 'com.miui.securitycenter',
    'oppo': 'com.coloros.safecenter',
    'realme': 'com.coloros.safecenter',
    'vivo': 'com.vivo.permissionmanager',
    'iqoo': 'com.vivo.permissionmanager',
    'huawei': 'com.huawei.systemmanager',
    'honor': 'com.huawei.systemmanager',
    'samsung': 'com.samsung.android.lool',
    'oneplus': 'com.oneplus.security',
    'meizu': 'com.meizu.safe',
    'zte': 'com.zte.heartyservice',
    'letv': 'com.letv.android.letvsafe',
    'lenovo': 'com.lenovo.security',
    'asus': 'com.asus.mobilemanager',
  };

  /// 检查是否是国产手机
  static Future<bool> isChinesePhone() async {
    if (!Platform.isAndroid) return false;

    final deviceInfo = DeviceInfoPlugin();
    final androidInfo = await deviceInfo.androidInfo;
    final manufacturer = androidInfo.manufacturer.toLowerCase();

    return _manufacturerIntents.keys.any(
      (key) => manufacturer.contains(key),
    );
  }

  /// 获取当前设备厂商名称
  static Future<String> getManufacturer() async {
    if (!Platform.isAndroid) return '';

    final deviceInfo = DeviceInfoPlugin();
    final androidInfo = await deviceInfo.androidInfo;
    return androidInfo.manufacturer;
  }

  /// 检查是否已经忽略电池优化
  static Future<bool> isIgnoringBatteryOptimizations() async {
    if (!Platform.isAndroid) return true;
    return await Permission.ignoreBatteryOptimizations.isGranted;
  }

  /// 请求忽略电池优化
  static Future<bool> requestIgnoreBatteryOptimizations() async {
    if (!Platform.isAndroid) return true;

    final status = await Permission.ignoreBatteryOptimizations.request();
    return status.isGranted;
  }

  /// 打开电池优化设置页面
  static Future<void> openBatteryOptimizationSettings() async {
    if (!Platform.isAndroid) return;

    try {
      await openAppSettings();
    } catch (e) {
      debugPrint('打开电池优化设置失败: $e');
    }
  }

  /// 尝试打开自启动设置页面
  static Future<bool> openAutoStartSettings() async {
    if (!Platform.isAndroid) return false;

    final deviceInfo = DeviceInfoPlugin();
    final androidInfo = await deviceInfo.androidInfo;
    final manufacturer = androidInfo.manufacturer.toLowerCase();

    String? packageName;
    for (var entry in _manufacturerIntents.entries) {
      if (manufacturer.contains(entry.key)) {
        packageName = entry.value;
        break;
      }
    }

    if (packageName != null) {
      try {
        // 使用 MethodChannel 打开特定应用
        const platform = MethodChannel('android_background_helper');
        final result = await platform.invokeMethod('openAutoStartSettings', {
          'packageName': packageName,
        });
        return result == true;
      } catch (e) {
        debugPrint('打开自启动设置失败: $e');
        // 失败时打开应用详情页
        await openAppSettings();
        return true;
      }
    }

    return false;
  }

  /// 显示后台权限引导对话框
  static Future<void> showBackgroundPermissionGuide({
    bool forceShow = false,
  }) async {
    if (!Platform.isAndroid) return;

    final prefs = await SharedPreferences.getInstance();
    final hasShown = prefs.getBool(_prefKeyShownAutoStart) ?? false;

    if (!forceShow && hasShown) return;

    final isChineseDevice = await isChinesePhone();
    if (!isChineseDevice && !forceShow) return;

    final manufacturer = await getManufacturer();

    Get.dialog(
      _BackgroundPermissionDialog(
        manufacturer: manufacturer,
        onAutoStartTap: () async {
          await openAutoStartSettings();
        },
        onBatteryTap: () async {
          await requestIgnoreBatteryOptimizations();
        },
        onDismiss: () async {
          await prefs.setBool(_prefKeyShownAutoStart, true);
        },
      ),
      barrierDismissible: false,
    );
  }

  /// 显示电池优化引导对话框
  static Future<void> showBatteryOptimizationGuide({
    bool forceShow = false,
  }) async {
    if (!Platform.isAndroid) return;

    final prefs = await SharedPreferences.getInstance();
    final hasShown = prefs.getBool(_prefKeyShownBattery) ?? false;

    if (!forceShow && hasShown) return;

    final isIgnoring = await isIgnoringBatteryOptimizations();
    if (isIgnoring && !forceShow) return;

    Get.dialog(
      _BatteryOptimizationDialog(
        onConfirm: () async {
          await requestIgnoreBatteryOptimizations();
          await prefs.setBool(_prefKeyShownBattery, true);
        },
        onCancel: () async {
          await prefs.setBool(_prefKeyShownBattery, true);
        },
      ),
      barrierDismissible: false,
    );
  }

  /// 检查并显示所有后台权限引导
  static Future<void> checkAndShowGuides() async {
    if (!Platform.isAndroid) return;

    // 首先检查电池优化
    final isIgnoring = await isIgnoringBatteryOptimizations();
    if (!isIgnoring) {
      await showBatteryOptimizationGuide();
      return;
    }

    // 然后检查自启动（仅国产手机）
    final isChineseDevice = await isChinesePhone();
    if (isChineseDevice) {
      await showBackgroundPermissionGuide();
    }
  }
}

/// 后台权限引导对话框
class _BackgroundPermissionDialog extends StatelessWidget {
  final String manufacturer;
  final VoidCallback onAutoStartTap;
  final VoidCallback onBatteryTap;
  final VoidCallback onDismiss;

  const _BackgroundPermissionDialog({
    required this.manufacturer,
    required this.onAutoStartTap,
    required this.onBatteryTap,
    required this.onDismiss,
  });

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12.r),
      ),
      title: Text(
        StrRes.backgroundPermissionTitle,
        style: TextStyle(
          fontSize: 18.sp,
          fontWeight: FontWeight.bold,
        ),
      ),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            StrRes.backgroundPermissionDesc,
            style: TextStyle(fontSize: 14.sp),
          ),
          SizedBox(height: 16.h),
          _buildPermissionItem(
            icon: Icons.battery_charging_full,
            title: StrRes.batteryOptimization,
            subtitle: StrRes.batteryOptimizationDesc,
            onTap: () {
              Get.back();
              onBatteryTap();
            },
          ),
          SizedBox(height: 12.h),
          _buildPermissionItem(
            icon: Icons.autorenew,
            title: StrRes.autoStart,
            subtitle: StrRes.autoStartDesc,
            onTap: () {
              Get.back();
              onAutoStartTap();
            },
          ),
          if (manufacturer.isNotEmpty) ...[
            SizedBox(height: 12.h),
            Text(
              '${StrRes.currentDevice}: $manufacturer',
              style: TextStyle(
                fontSize: 12.sp,
                color: Colors.grey,
              ),
            ),
          ],
        ],
      ),
      actions: [
        TextButton(
          onPressed: () {
            onDismiss();
            Get.back();
          },
          child: Text(
            StrRes.notNow,
            style: const TextStyle(color: Colors.grey),
          ),
        ),
      ],
    );
  }

  Widget _buildPermissionItem({
    required IconData icon,
    required String title,
    required String subtitle,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8.r),
      child: Container(
        padding: EdgeInsets.all(12.w),
        decoration: BoxDecoration(
          color: Colors.grey.withOpacity(0.1),
          borderRadius: BorderRadius.circular(8.r),
        ),
        child: Row(
          children: [
            Icon(icon, size: 32.w, color: Styles.c_0089FF),
            SizedBox(width: 12.w),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: TextStyle(
                      fontSize: 14.sp,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  SizedBox(height: 2.h),
                  Text(
                    subtitle,
                    style: TextStyle(
                      fontSize: 12.sp,
                      color: Colors.grey,
                    ),
                  ),
                ],
              ),
            ),
            const Icon(Icons.chevron_right, color: Colors.grey),
          ],
        ),
      ),
    );
  }
}

/// 电池优化引导对话框
class _BatteryOptimizationDialog extends StatelessWidget {
  final VoidCallback onConfirm;
  final VoidCallback onCancel;

  const _BatteryOptimizationDialog({
    required this.onConfirm,
    required this.onCancel,
  });

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12.r),
      ),
      title: Row(
        children: [
          Icon(Icons.battery_alert, color: Colors.orange, size: 24.w),
          SizedBox(width: 8.w),
          Text(
            StrRes.batteryOptimizationTitle,
            style: TextStyle(
              fontSize: 18.sp,
              fontWeight: FontWeight.bold,
            ),
          ),
        ],
      ),
      content: Text(
        StrRes.batteryOptimizationContent,
        style: TextStyle(fontSize: 14.sp),
      ),
      actions: [
        TextButton(
          onPressed: () {
            onCancel();
            Get.back();
          },
          child: Text(
            StrRes.notNow,
            style: const TextStyle(color: Colors.grey),
          ),
        ),
        ElevatedButton(
          onPressed: () {
            Get.back();
            onConfirm();
          },
          style: ElevatedButton.styleFrom(
            backgroundColor: Styles.c_0089FF,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(8.r),
            ),
          ),
          child: Text(
            StrRes.goToSettings,
            style: const TextStyle(color: Colors.white),
          ),
        ),
      ],
    );
  }
}
