import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../models.dart';
import '../api.dart';
import 'package:cached_network_image/cached_network_image.dart';

class FileItemWidget extends StatelessWidget {
  final FileItem item;
  final String layout;
  final ApiService api;
  final VoidCallback onTap;

  const FileItemWidget({
    Key? key,
    required this.item,
    required this.layout,
    required this.api,
    required this.onTap,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    if (layout == 'grid') {
      return _buildGridItem(context);
    } else {
      return _buildListItem(context, layout == 'details');
    }
  }

  Widget _buildGridItem(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Card(
      elevation: 0,
      color: cs.surfaceVariant,
      child: InkWell(
        onTap: onTap,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Expanded(
              child: _buildIconOrThumb(context, size: 100),
            ),
            Padding(
              padding: const EdgeInsets.all(8.0),
              child: Text(
                item.name,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.labelMedium,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildListItem(BuildContext context, bool showDetails) {
    final cs = Theme.of(context).colorScheme;
    return ListTile(
      leading: Container(
        width: 48,
        height: 48,
        decoration: BoxDecoration(
          color: cs.surfaceVariant,
          borderRadius: BorderRadius.circular(8),
        ),
        clipBehavior: Clip.antiAlias,
        child: _buildIconOrThumb(context, size: 40),
      ),
      title: Text(item.name, maxLines: 1, overflow: TextOverflow.ellipsis),
      subtitle: showDetails ? Text(_formatDetails()) : null,
      onTap: onTap,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
    );
  }

  String _formatDetails() {
    if (item.isDir) return 'Directory';
    final size = _formatSize(item.size);
    final date = DateFormat.yMMMd().format(DateTime.fromMillisecondsSinceEpoch(item.modTime * 1000));
    return '$size • $date';
  }

  String _formatSize(int bytes) {
    if (bytes <= 0) return "0 B";
    const suffixes = ["B", "KB", "MB", "GB", "TB"];
    var i = 0;
    double size = bytes.toDouble();
    while (size >= 1024 && i < suffixes.length - 1) {
      size /= 1024;
      i++;
    }
    return "${size.toStringAsFixed(1)} ${suffixes[i]}";
  }

  Widget _buildIconOrThumb(BuildContext context, {required double size}) {
    if (item.canRender && !item.name.toLowerCase().endsWith('.pdf')) {
      return CachedNetworkImage(
        imageUrl: api.getThumbUrl(item.path),
        placeholder: (context, url) => const Center(child: CircularProgressIndicator(strokeWidth: 2)),
        errorWidget: (context, url, error) => _buildDefaultIcon(context),
        fit: BoxFit.cover,
      );
    }
    return _buildDefaultIcon(context);
  }

  Widget _buildDefaultIcon(BuildContext context) {
    IconData iconData;
    final cs = Theme.of(context).colorScheme;

    if (item.isDir) {
      iconData = Icons.folder;
      // primary for folders
      return Icon(iconData, color: cs.primary, size: layout == 'grid' ? 40 : 24);
    } else if (item.name.toLowerCase().endsWith('.zip')) {
      iconData = Icons.archive;
      return Icon(iconData, color: cs.secondary, size: layout == 'grid' ? 40 : 24);
    } else if (item.canStream) {
      iconData = Icons.movie;
      return Icon(iconData, color: cs.tertiary, size: layout == 'grid' ? 40 : 24);
    } else if (item.canRender) {
      iconData = item.name.toLowerCase().endsWith('.pdf') ? Icons.picture_as_pdf : Icons.image;
      return Icon(iconData, color: cs.secondary, size: layout == 'grid' ? 40 : 24);
    } else if (item.canEdit) {
      iconData = Icons.description;
      return Icon(iconData, color: cs.primaryContainer, size: layout == 'grid' ? 40 : 24);
    } else {
      iconData = Icons.insert_drive_file;
      return Icon(iconData, color: cs.outline, size: layout == 'grid' ? 40 : 24);
    }
  }
}
