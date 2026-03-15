import 'package:flutter/material.dart';

class BreadcrumbsWidget extends StatelessWidget {
  final String path;
  final Function(String) onNavigate;

  const BreadcrumbsWidget({
    Key? key,
    required this.path,
    required this.onNavigate,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    final parts = path.split('/').where((p) => p.isNotEmpty).toList();
    
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      padding: const EdgeInsets.symmetric(horizontal: 8),
      child: Row(
        children: [
          IconButton(
            icon: const Icon(Icons.home),
            onPressed: () => onNavigate('/'),
            visualDensity: VisualDensity.compact,
          ),
          ...parts.asMap().entries.map((entry) {
            final index = entry.key;
            final part = entry.value;
            final fullPath = '/' + parts.sublist(0, index + 1).join('/');
            
            return Row(
              children: [
                const Icon(Icons.chevron_right, size: 16, color: Colors.grey),
                TextButton(
                  onPressed: () => onNavigate(fullPath),
                  child: Text(part),
                  style: TextButton.styleFrom(
                    visualDensity: VisualDensity.compact,
                    padding: const EdgeInsets.symmetric(horizontal: 4),
                  ),
                ),
              ],
            );
          }).toList(),
        ],
      ),
    );
  }
}
