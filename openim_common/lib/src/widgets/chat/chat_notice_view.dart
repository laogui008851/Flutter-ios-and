import 'package:flutter/material.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';

class ChatNoticeView extends StatelessWidget {
  final bool isISend;
  final String content;
  
  const ChatNoticeView({
    Key? key,
    required this.isISend,
    required this.content,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.symmetric(
        horizontal: 12.w,
        vertical: 8.h,
      ),
      decoration: BoxDecoration(
        color: isISend 
          ? const Color(0xFF1B72EC).withOpacity(0.1)
          : const Color(0xFFF0F0F0),
        borderRadius: BorderRadius.circular(6.r),
      ),
      child: Text(
        content,
        style: TextStyle(
          fontSize: 14.sp,
          color: isISend 
            ? const Color(0xFF1B72EC)
            : const Color(0xFF333333),
        ),
      ),
    );
  }
}