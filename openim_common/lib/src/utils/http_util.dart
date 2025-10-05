import 'dart:io';
import 'dart:ui';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_easyloading/flutter_easyloading.dart';
import 'package:image_gallery_saver_plus/image_gallery_saver_plus.dart';
import 'package:openim_common/openim_common.dart';
import 'package:talker_dio_logger/talker_dio_logger.dart';

var dio = Dio();

class HttpUtil {
  HttpUtil._();

  static void init() {
    dio
      ..interceptors.add(
        TalkerDioLogger(
          settings: const TalkerDioLoggerSettings(
            printRequestHeaders: kDebugMode,
            printRequestData: kDebugMode,
            printResponseMessage: kDebugMode,
            printResponseData: kDebugMode,
            printResponseHeaders: kDebugMode,
          ),
        ),
      )
      ..interceptors.add(InterceptorsWrapper(onRequest: (options, handler) {
        return handler.next(options); //continue
      }, onResponse: (response, handler) {
        return handler.next(response); // continue
      }, onError: (DioError e, handler) {
        return handler.next(e); //continue
      }));

    dio.options.baseUrl = Config.imApiUrl;
    dio.options.connectTimeout = const Duration(seconds: 30); //30s
    dio.options.receiveTimeout = const Duration(seconds: 30);
  }

  static String get operationID => DateTime.now().millisecondsSinceEpoch.toString();

  static Future post(
    String path, {
    dynamic data,
    bool showErrorToast = true,
    Map<String, dynamic>? queryParameters,
    Options? options,
    CancelToken? cancelToken,
    ProgressCallback? onSendProgress,
    ProgressCallback? onReceiveProgress,
  }) async {
    try {
      data ??= {};
      options ??= Options();
      options.headers ??= {};
      options.headers!['operationID'] = operationID;

      var result = await dio.post<Map<String, dynamic>>(
        path,
        data: data,
        queryParameters: queryParameters,
        options: options,
        cancelToken: cancelToken,
        onSendProgress: onSendProgress,
        onReceiveProgress: onReceiveProgress,
      );
      var resp = ApiResp.fromJson(result.data!);
      if (resp.errCode == 0) {
        return resp.data;
      } else {
        if (showErrorToast) {
          // 对后端返回的错误信息进行本地化处理
          final localizedError = _localizeErrorMessage(resp.errDlt, resp.errMsg, resp.errCode);
          IMViews.showToast(localizedError);
        }

        return Future.error((resp.errCode, resp.errMsg));
      }
    } catch (error) {
      if (error is DioException) {
        // 过滤敏感信息，避免在正式包中暴露接口地址
        final filteredError = _filterSensitiveInfo('接口：$path  信息：${error.message}');
        if (showErrorToast) IMViews.showToast(filteredError);
        return Future.error(filteredError);
      }
      final filteredError = _filterSensitiveInfo('接口：$path  信息：${error.toString()}');
      if (showErrorToast) IMViews.showToast(filteredError);
      return Future.error(filteredError);
    }
  }

  static Future<String> uploadImageForMinio({
    required String path,
    bool compress = true,
  }) async {
    String fileName = path.substring(path.lastIndexOf("/") + 1);

    String? compressPath;
    if (compress) {
      File? compressFile = await IMUtils.compressImageAndGetFile(File(path));
      compressPath = compressFile?.path;
      Logger.print('compressPath: $compressPath');
    }
    final bytes = await File(compressPath ?? path).readAsBytes();
    final mf = MultipartFile.fromBytes(bytes, filename: fileName);

    var formData =
        FormData.fromMap({'operationID': '${DateTime.now().millisecondsSinceEpoch}', 'fileType': 1, 'file': mf});

    var resp = await dio.post<Map<String, dynamic>>(
      "${Config.imApiUrl}/third/minio_upload",
      data: formData,
      options: Options(headers: {'token': DataSp.imToken}),
    );
    return resp.data?['data']['URL'];
  }

  static Future download(
    String url, {
    required String cachePath,
    CancelToken? cancelToken,
    Function(int count, int total)? onProgress,
  }) {
    return dio.download(
      url,
      cachePath,
      options: Options(
        receiveTimeout: const Duration(minutes: 10),
      ),
      cancelToken: cancelToken,
      onReceiveProgress: onProgress,
    );
  }

  static Future saveUrlPicture(
    String url, {
    CancelToken? cancelToken,
    Function(int count, int total)? onProgress,
    VoidCallback? onCompletion,
  }) async {
    final name = url.substring(url.lastIndexOf('/') + 1);
    final cachePath = await IMUtils.createTempFile(dir: 'picture', name: name);
    var intervalDo = IntervalDo();

    return download(
      url,
      cachePath: cachePath,
      cancelToken: cancelToken,
      onProgress: (int count, int total) async {
        onProgress?.call(count, total);
        if (total == -1) {
          onCompletion?.call();
          intervalDo.drop(
              fun: () async {
                saveFileToGallerySaver(File(cachePath), showTaost: EasyLoading.isShow);
              },
              milliseconds: 1500);
        }
        if (count == total) {
          saveFileToGallerySaver(File(cachePath), showTaost: EasyLoading.isShow);
        }
      },
    );
  }

  static Future saveImage(Image image) async {
    var byteData = await image.toByteData(format: ImageByteFormat.png);
    if (byteData != null) {
      Uint8List uint8list = byteData.buffer.asUint8List();
      var result = await ImageGallerySaverPlus.saveImage(Uint8List.fromList(uint8list));
      if (result != null) {
        var tips = StrRes.saveSuccessfully;
        if (Platform.isAndroid) {
          final filePath = result['filePath'].split('//').last;
          tips = '${StrRes.saveSuccessfully}:$filePath';
        }
        IMViews.showToast(tips);
      }
    }
  }

  static Future saveUrlVideo(
    String url, {
    CancelToken? cancelToken,
    Function(int count, int total)? onProgress,
    VoidCallback? onCompletion,
  }) async {
    final name = url.substring(url.lastIndexOf('/') + 1);
    final cachePath = await IMUtils.createTempFile(dir: 'video', name: name);

    if (File(cachePath).existsSync()) {
      onCompletion?.call();
      return;
    }

    return download(
      url,
      cachePath: cachePath,
      cancelToken: cancelToken,
      onProgress: (int count, int total) async {
        onProgress?.call(count, total);
        if (count == total) {
          onCompletion?.call();
          final result = await ImageGallerySaverPlus.saveFile(cachePath);
          if (result != null) {
            var tips = StrRes.saveSuccessfully;
            if (Platform.isAndroid) {
              final filePath = result['filePath'].split('//').last;
              tips = '${StrRes.saveSuccessfully}:$filePath';
            }
            IMViews.showToast(tips);
          }
        }
      },
    );
  }

  static Future saveFileToGallerySaver(File file, {String? name, bool showTaost = true}) async {
    Permissions.storage(() async {
      var tips = StrRes.saveSuccessfully;
      Logger.print('saveFileToGallerySaver: ${file.path}');
      final imageBytes = await file.readAsBytes();

      final result = await ImageGallerySaverPlus.saveImage(imageBytes, name: name);
      if (result != null && showTaost) {
        if (Platform.isAndroid) {
          final filePath = result['filePath'].split('//').last;
          tips = '${StrRes.saveSuccessfully}:$filePath';
        }
        IMViews.showToast(tips);
      }
    });
  }
  
  /// 过滤敏感信息，如IP地址、端口等
  static String _filterSensitiveInfo(String message) {
    // 对于网络连接错误，直接返回友好的中文提示
    if (message.toLowerCase().contains('connection') || 
        message.toLowerCase().contains('network') ||
        message.toLowerCase().contains('timeout') ||
        message.toLowerCase().contains('failed') ||
        message.toLowerCase().contains('error')) {
      return '网络连接失败，请检查网络设置';
    }
    
    // 移除IP地址和端口信息
    String filtered = message
        // 首先匹配完整的HTTP URL（包含路径）
        .replaceAll(RegExp(r'https?://[^\s/]+/[^\s]*'), '[服务器接口]')
        // 匹配带端口的HTTP URL
        .replaceAll(RegExp(r'https?://[^\s/]+:[0-9]+'), '[服务器地址]')
        // 匹配IPv4地址:端口 格式
        .replaceAll(RegExp(r'\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}:[0-9]+\b'), '[服务器地址]')
        // 匹配IPv4地址
        .replaceAll(RegExp(r'\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b'), '[服务器地址]')
        // 匹配所有HTTP/HTTPS URL
        .replaceAll(RegExp(r'https?://[^\s]+'), '[服务器地址]')
        // 移除英文技术错误信息
        .replaceAll(RegExp(r'The connection errored:.*'), '')
        .replaceAll(RegExp(r'Connection failed.*'), '')
        .replaceAll(RegExp(r'This indicates an error.*'), '')
        .replaceAll(RegExp(r'cannot be solved by the library.*'), '')
        .replaceAll(RegExp(r'SocketException.*'), '')
        .replaceAll(RegExp(r'TimeoutException.*'), '')
        // 匹配单独的端口号
        .replaceAll(RegExp(r':(\d{4,5})'), '')
        // 移除具体的域名
        .replaceAll(RegExp(r'[a-zA-Z0-9-]+\.[a-zA-Z]{2,}'), '[服务器地址]')
        // 移除具体的接口路径
        .replaceAll(RegExp(r'/account/[^\s]*'), '/[接口路径]')
        .replaceAll(RegExp(r'/[a-zA-Z_/]+'), '/[接口路径]');
    
    // 如果过滤后为空或只剩下无用信息，返回友好的错误信息
    if (filtered.trim().isEmpty || 
        filtered.trim() == '接口：[服务器接口] 信息：' ||
        filtered.contains('接口：') && filtered.replaceAll(RegExp(r'接口：.*信息：'), '').trim().isEmpty) {
      return '网络连接失败，请检查网络设置';
    }
    
    return filtered;
  }
  
  /// 本地化错误信息
  static String _localizeErrorMessage(String errDlt, String errMsg, int errCode) {
    // 常见的英文错误信息本地化映射
    final String message = errDlt.isNotEmpty ? errDlt : errMsg;
    
    // 账户相关错误
    if (message.toLowerCase().contains('accountnotfound') || 
        message.toLowerCase().contains('account not found') ||
        message.toLowerCase().contains('user not found')) {
      return '账户不存在，请检查手机号或邮箱';
    }
    
    if (message.toLowerCase().contains('password') && 
        (message.toLowerCase().contains('wrong') || message.toLowerCase().contains('incorrect'))) {
      return '密码错误，请重新输入';
    }
    
    if (message.toLowerCase().contains('verification code') ||
        message.toLowerCase().contains('code') && message.toLowerCase().contains('invalid')) {
      return '验证码错误或已过期';
    }
    
    // 网络相关错误
    if (message.toLowerCase().contains('connection') || 
        message.toLowerCase().contains('network') ||
        message.toLowerCase().contains('timeout')) {
      return '网络连接失败，请检查网络设置';
    }
    
    // 权限相关错误
    if (message.toLowerCase().contains('permission') || 
        message.toLowerCase().contains('unauthorized') ||
        message.toLowerCase().contains('forbidden')) {
      return '权限不足，请重新登录';
    }
    
    // 服务器相关错误
    if (message.toLowerCase().contains('server') || 
        message.toLowerCase().contains('internal')) {
      return '服务器繁忙，请稍后重试';
    }
    
    // 根据错误码进行本地化
    switch (errCode) {
      case 1001:
      case 1002:
        return '账户信息错误';
      case 1003:
      case 1004:
        return '验证码错误';
      case 1005:
        return '账户已存在';
      case 2001:
      case 2002:
        return '网络连接失败';
      case 3001:
        return '权限不足';
      case 5001:
      case 5002:
        return '服务器错误';
      default:
        // 如果无法识别，返回通用友好提示，避免显示英文技术错误
        return message.length > 50 ? '操作失败，请稍后重试' : message;
    }
  }
}
