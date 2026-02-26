package cn.rentsoft.flutter.openim.consumer;

import android.content.ComponentName;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.os.Bundle;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import io.flutter.embedding.android.FlutterFragmentActivity;
import io.flutter.embedding.engine.FlutterEngine;
import io.flutter.plugin.common.MethodChannel;

public class MainActivity extends FlutterFragmentActivity {
    private static final String CHANNEL = "android_background_helper";

    @Override
    public void configureFlutterEngine(@NonNull FlutterEngine flutterEngine) {
        super.configureFlutterEngine(flutterEngine);

        new MethodChannel(flutterEngine.getDartExecutor().getBinaryMessenger(), CHANNEL)
                .setMethodCallHandler((call, result) -> {
                    if (call.method.equals("openAutoStartSettings")) {
                        String packageName = call.argument("packageName");
                        boolean success = openAutoStartSettings(packageName);
                        result.success(success);
                    } else {
                        result.notImplemented();
                    }
                });
    }

    private boolean openAutoStartSettings(String packageName) {
        if (packageName == null) {
            return false;
        }

        try {
            // 根据不同厂商尝试打开自启动设置
            Intent intent = new Intent();
            intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);

            switch (packageName) {
                case "com.miui.securitycenter":
                    // 小米
                    intent.setComponent(new ComponentName("com.miui.securitycenter",
                            "com.miui.permcenter.autostart.AutoStartManagementActivity"));
                    break;
                case "com.coloros.safecenter":
                    // OPPO/Realme
                    intent.setComponent(new ComponentName("com.coloros.safecenter",
                            "com.coloros.safecenter.permission.startup.StartupAppListActivity"));
                    break;
                case "com.vivo.permissionmanager":
                    // vivo/iQOO
                    intent.setComponent(new ComponentName("com.vivo.permissionmanager",
                            "com.vivo.permissionmanager.activity.BgStartUpManagerActivity"));
                    break;
                case "com.huawei.systemmanager":
                    // 华为/荣耀
                    intent.setComponent(new ComponentName("com.huawei.systemmanager",
                            "com.huawei.systemmanager.startupmgr.ui.StartupNormalAppListActivity"));
                    break;
                case "com.samsung.android.lool":
                    // 三星
                    intent.setComponent(new ComponentName("com.samsung.android.lool",
                            "com.samsung.android.sm.battery.ui.BatteryActivity"));
                    break;
                case "com.oneplus.security":
                    // 一加
                    intent.setComponent(new ComponentName("com.oneplus.security",
                            "com.oneplus.security.chainlaunch.view.ChainLaunchAppListActivity"));
                    break;
                case "com.meizu.safe":
                    // 魅族
                    intent.setComponent(new ComponentName("com.meizu.safe",
                            "com.meizu.safe.permission.SmartBGActivity"));
                    break;
                default:
                    // 其他厂商，尝试打开应用详情页
                    return openAppDetailSettings();
            }

            // 检查目标 Activity 是否存在
            PackageManager pm = getPackageManager();
            if (intent.resolveActivity(pm) != null) {
                startActivity(intent);
                return true;
            } else {
                // 如果特定页面不存在，尝试打开应用详情页
                return openAppDetailSettings();
            }
        } catch (Exception e) {
            e.printStackTrace();
            // 发生异常时尝试打开应用详情页
            return openAppDetailSettings();
        }
    }

    private boolean openAppDetailSettings() {
        try {
            Intent intent = new Intent(android.provider.Settings.ACTION_APPLICATION_DETAILS_SETTINGS);
            intent.setData(android.net.Uri.parse("package:" + getPackageName()));
            intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
            startActivity(intent);
            return true;
        } catch (Exception e) {
            e.printStackTrace();
            return false;
        }
    }
}
