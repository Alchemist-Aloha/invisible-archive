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
    
    final cs = Theme.of(context).colorScheme;

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      padding: const EdgeInsets.symmetric(horizontal: 8),
      child: Row(
        children: [
          IconButton(
            icon: Icon(Icons.home, color: cs.primary),
            onPressed: () => onNavigate('/'),
            visualDensity: VisualDensity.compact,
            tooltip: 'Home',
          ),
          ...parts.asMap().entries.map((entry) {
            final index = entry.key;
            final part = entry.value;
            final fullPath = '/' + parts.sublist(0, index + 1).join('/');

            return Row(
              children: [
                Icon(Icons.chevron_right, size: 16, color: cs.onSurfaceVariant),
                TextButton(
                  onPressed: () => onNavigate(fullPath),
                  child: Text(part),
                  style: TextButton.styleFrom(
                    visualDensity: VisualDensity.compact,
                    padding: const EdgeInsets.symmetric(horizontal: 4),
                    foregroundColor: cs.primary,
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
