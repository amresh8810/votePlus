import { BarChart3, Download, FileText, Printer, QrCode, Share2 } from 'lucide-react';
import { downloadCsv, downloadQrCode, formatPublicPollUrl, shareOrCopy } from '../utils/exportUtils';

export const PollExportActions = ({ poll, counts, totalVotes }) => {
  const publicUrl = formatPublicPollUrl(poll.id);

  const exportResults = () => {
    downloadCsv(
      `${poll.question.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '') || 'poll'}-results.csv`,
      ['Option', 'Votes', 'Percentage'],
      (poll.options || []).map((option) => {
        const votes = counts[option.id] || 0;
        const percentage = totalVotes ? ((votes / totalVotes) * 100).toFixed(1) : '0.0';
        return [option.text, votes, `${percentage}%`];
      }),
    );
  };

  const sharePoll = () => shareOrCopy(poll.question, publicUrl);

  return (
    <div className="poll-export-actions" aria-label="Export and share actions">
      <button type="button" className="btn btn-outline btn-sm" onClick={exportResults}>
        <Download size={14} />
        <span>Export CSV</span>
      </button>
      <button type="button" className="btn btn-outline btn-sm" onClick={() => window.print()}>
        <Printer size={14} />
        <span>Print Results</span>
      </button>
      <button type="button" className="btn btn-outline btn-sm" onClick={sharePoll}>
        <Share2 size={14} />
        <span>Share</span>
      </button>
      <a
        className="btn btn-outline btn-sm"
        href={`https://quickchart.io/qr?text=${encodeURIComponent(publicUrl)}&size=220`}
        target="_blank"
        rel="noopener noreferrer"
      >
        <QrCode size={14} />
        <span>Open QR</span>
      </a>
      <button type="button" className="btn btn-outline btn-sm" onClick={() => downloadQrCode(publicUrl, 'poll-qr-code.png')}>
        <QrCode size={14} />
        <span>Download QR</span>
      </button>
    </div>
  );
};

export const AnalyticsExportActions = ({ points, period }) => {
  const exportAnalytics = () => {
    downloadCsv(
      `votes-over-time-${period}.csv`,
      ['Date', 'Votes'],
      points.map((point) => [point.date, point.count]),
    );
  };

  return (
    <div className="analytics-export-actions" aria-label="Analytics export actions">
      <button type="button" className="btn btn-outline btn-sm" onClick={exportAnalytics}>
        <BarChart3 size={14} />
        <span>Download analytics</span>
      </button>
      <button type="button" className="btn btn-outline btn-sm" onClick={() => window.print()}>
        <FileText size={14} />
        <span>PDF / Print</span>
      </button>
    </div>
  );
};
