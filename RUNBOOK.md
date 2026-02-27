# FunctionFly Content Management Runbook

## Overview

This runbook provides procedures for maintaining and troubleshooting the FunctionFly content management system, which includes Sanity Studio for content editing and the FunctionFly dashboard for operational management.

## System Architecture

- **Content Management**: Sanity Studio with custom schemas
- **Site Generation**: Astro site with Sanity integration
- **Admin Dashboard**: React dashboard for operational tasks
- **Hosting**: Vercel (site + studio), separate hosting for dashboard
- **CDN**: Sanity CDN for content, Vercel CDN for site assets

## Monitoring & Health Checks

### Health Endpoints

- Site health: `GET /api/health`
- Dashboard health: Check dashboard uptime (TBD)

### Key Metrics to Monitor

- Webhook processing success rate
- Content publishing latency
- Redirect generation success
- Newsletter signup conversion
- API response times

### Alert Conditions

- Webhook signature verification failures
- Content publishing failures
- Rate limit exceeded (webhooks)
- Sanity API timeouts
- Memory usage > 80%

## Common Issues & Solutions

### Issue: Content Not Appearing on Live Site

**Symptoms:**

- Content shows in Sanity Studio but not on website
- Recent changes not reflected

**Diagnosis:**

1. Check content status in Sanity (must be "published")
2. Verify build/deployment status
3. Check Sanity API connectivity: `GET /api/health`

**Solutions:**

1. Ensure content status is "published" in Studio
2. Trigger manual rebuild if needed
3. Check environment variables (SANITY_PROJECT_ID, etc.)
4. Verify webhook delivery if using automated publishing

### Issue: Webhook Processing Failures

**Symptoms:**

- Alert: "Webhook processing failed"
- Content changes not triggering expected actions

**Diagnosis:**

1. Check webhook signature verification
2. Verify SANITY_AUTH_TOKEN permissions
3. Check rate limiting status
4. Review webhook event logs in Sanity

**Solutions:**

1. Verify SANITY_WEBHOOK_SECRET environment variable
2. Ensure webhook endpoint has proper permissions
3. Check webhook payload format
4. Implement retry logic for transient failures

### Issue: Redirects Not Working

**Symptoms:**

- 404 errors for expected redirects
- Old URLs still showing

**Diagnosis:**

1. Check redirect status in Sanity (must be enabled)
2. Verify _redirects file generation
3. Check hosting platform redirect configuration

**Solutions:**

1. Run `npm run generate-redirects` manually
2. Ensure redirect has valid source/destination
3. Check build logs for redirect generation errors
4. Verify hosting platform supports _redirects format

### Issue: Newsletter Signup Failures

**Symptoms:**

- Signup forms not working
- Subscribers not appearing in admin

**Diagnosis:**

1. Check newsletter provider API status
2. Verify API keys and credentials
3. Check form validation
4. Review error logs

**Solutions:**

1. Verify newsletter provider configuration
2. Check API rate limits
3. Implement fallback for provider outages
4. Add form validation feedback

### Issue: Studio Performance Issues

**Symptoms:**

- Slow loading in Sanity Studio
- Timeout errors

**Diagnosis:**

1. Check Sanity API performance
2. Verify network connectivity
3. Check browser console for errors
4. Review Studio configuration

**Solutions:**

1. Optimize GROQ queries
2. Implement query caching where appropriate
3. Check for large media files
4. Verify Studio plugins are up to date

## Emergency Procedures

### Complete Site Outage

1. **Assess Impact**: Determine affected components (site, studio, dashboard)
2. **Check Dependencies**: Verify Sanity, hosting providers status
3. **Rollback**: If recent deployment caused issues, rollback to previous version
4. **Communication**: Notify team and users via status page/social media
5. **Recovery**: Restore from backups if data loss occurred

### Data Loss Incident

1. **Stop Changes**: Prevent further content modifications
2. **Assess Scope**: Determine what data was lost
3. **Restore from Backup**: Use Sanity's backup/restore features
4. **Verify Integrity**: Check restored content for completeness
5. **Document**: Record incident details for post-mortem

### Security Incident

1. **Isolate**: Disconnect affected systems
2. **Assess Breach**: Determine what was compromised
3. **Change Credentials**: Rotate all API keys and passwords
4. **Notify**: Inform affected parties and authorities if required
5. **Audit**: Review access logs for suspicious activity

## Maintenance Procedures

### Regular Tasks

**Daily:**

- Check health endpoints
- Review error logs
- Monitor webhook processing

**Weekly:**

- Review content publishing queue
- Check redirect effectiveness
- Audit user permissions

**Monthly:**

- Update dependencies
- Review and rotate API keys
- Backup verification
- Performance optimization

### Content Publishing Workflow

1. **Draft Creation**: Author creates content in Studio
2. **Review Process**: Content moves to "in_review" status
3. **Approval**: Editor approves and schedules or publishes
4. **Publishing**: Automated via webhooks or manual triggers
5. **Verification**: Check live site and monitoring alerts

### Deployment Process

1. **Pre-deployment**: Run health checks and tests
2. **Deploy**: Push changes to hosting platforms
3. **Verification**: Check all endpoints and functionality
4. **Monitoring**: Watch for errors in first 30 minutes
5. **Rollback Plan**: Have previous version ready if needed

## Contact Information

**Development Team:**

- Primary: [Primary Contact]
- Secondary: [Secondary Contact]

**Hosting Providers:**

- Vercel: [Support Contact]
- Sanity: [Support Contact]

**External Services:**

- Newsletter Provider: [Support Contact]
- Monitoring: [Service Contact]

## References

- [Sanity Documentation](https://www.sanity.io/docs)
- [Astro Documentation](https://docs.astro.build)
- [Vercel Documentation](https://vercel.com/docs)
- [Project Architecture](./plans/BLOG_ADMIN_PROJECT_PLAN.md)
