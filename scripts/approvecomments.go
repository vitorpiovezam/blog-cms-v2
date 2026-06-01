// approve-comments: interactive CLI to review and approve pending blog comments from DynamoDB.
//
// Usage:
//   go run ./scripts/approve-comments [--table blog-comments] [--region us-east-1] [--profile default]
//
// For each pending (active=false) comment it shows a preview and asks:
//   [A]pprove  [D]elete  [S]kip  [Q]uit

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// ── flags ────────────────────────────────────────────────────────────────────

var (
	flagTable   = flag.String("table", envOr("COMMENTS_TABLE", "blog-comments"), "DynamoDB table name")
	flagRegion  = flag.String("region", envOr("AWS_REGION", "us-east-1"), "AWS region")
	flagProfile = flag.String("profile", envOr("AWS_PROFILE", "default"), "AWS credentials profile")
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ── comment shape (mirrors definitions.Comment) ───────────────────────────

type Comment struct {
	Slug       string    `dynamodbav:"slug"`
	CommentID  string    `dynamodbav:"commentId"`
	Author     string    `dynamodbav:"author"`
	Text       string    `dynamodbav:"text"`
	CreatedAt  time.Time `dynamodbav:"createdAt"`
	Active     bool      `dynamodbav:"active"`
	ParentID   string    `dynamodbav:"parentId,omitempty"`
	Recommends int       `dynamodbav:"recommends"`
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	flag.Parse()

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(*flagRegion),
		config.WithSharedConfigProfile(*flagProfile),
	)
	if err != nil {
		fatalf("loading AWS config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	pending, err := fetchPending(ctx, client)
	if err != nil {
		fatalf("fetching comments: %v", err)
	}

	if len(pending) == 0 {
		fmt.Println("✅  No pending comments — all caught up!")
		return
	}

	fmt.Printf("📬  Found %d pending comment(s)\n\n", len(pending))

	reader := bufio.NewReader(os.Stdin)
	approved, deleted, skipped := 0, 0, 0

	for i, c := range pending {
		printComment(i+1, len(pending), c)

		for {
			fmt.Print("  → [A]pprove  [D]elete  [S]kip  [Q]uit: ")
			line, _ := reader.ReadString('\n')
			choice := strings.ToUpper(strings.TrimSpace(line))

			switch choice {
			case "A":
				if err := approve(ctx, client, c); err != nil {
					fmt.Printf("  ⚠️  Error approving: %v\n", err)
				} else {
					fmt.Println("  ✅  Approved")
					approved++
				}
				goto next
			case "D":
				if err := deleteComment(ctx, client, c); err != nil {
					fmt.Printf("  ⚠️  Error deleting: %v\n", err)
				} else {
					fmt.Println("  🗑️  Deleted")
					deleted++
				}
				goto next
			case "S", "":
				fmt.Println("  ⏭️  Skipped")
				skipped++
				goto next
			case "Q":
				printSummary(approved, deleted, skipped, len(pending)-i-1)
				os.Exit(0)
			default:
				fmt.Println("  Unknown choice. Press A, D, S or Q.")
			}
		}
	next:
		fmt.Println()
	}

	printSummary(approved, deleted, skipped, 0)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func fetchPending(ctx context.Context, client *dynamodb.Client) ([]Comment, error) {
	// Scan entire table for items where active = false
	out, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(*flagTable),
		FilterExpression: aws.String("#active = :false"),
		ExpressionAttributeNames: map[string]string{
			"#active": "active",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":false": &types.AttributeValueMemberBOOL{Value: false},
		},
	})
	if err != nil {
		return nil, err
	}

	var comments []Comment
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func approve(ctx context.Context, client *dynamodb.Client, c Comment) error {
	_, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(*flagTable),
		Key: map[string]types.AttributeValue{
			"slug":      &types.AttributeValueMemberS{Value: c.Slug},
			"commentId": &types.AttributeValueMemberS{Value: c.CommentID},
		},
		UpdateExpression: aws.String("SET #active = :true"),
		ExpressionAttributeNames: map[string]string{
			"#active": "active",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":true": &types.AttributeValueMemberBOOL{Value: true},
		},
	})
	return err
}

func deleteComment(ctx context.Context, client *dynamodb.Client, c Comment) error {
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(*flagTable),
		Key: map[string]types.AttributeValue{
			"slug":      &types.AttributeValueMemberS{Value: c.Slug},
			"commentId": &types.AttributeValueMemberS{Value: c.CommentID},
		},
	})
	return err
}

func printComment(n, total int, c Comment) {
	fmt.Printf("──────────────────────────────────────────────────\n")
	fmt.Printf("  Comment %d/%d\n", n, total)
	fmt.Printf("  Post   : %s\n", c.Slug)
	fmt.Printf("  Author : %s\n", c.Author)
	fmt.Printf("  Date   : %s\n", c.CreatedAt.Format("2006-01-02 15:04"))
	if c.ParentID != "" {
		fmt.Printf("  Reply to: %s\n", c.ParentID)
	}
	fmt.Printf("\n  %s\n\n", wrap(c.Text, 60))
}

func wrap(s string, width int) string {
	words := strings.Fields(s)
	var lines []string
	line := ""
	for _, w := range words {
		if len(line)+len(w)+1 > width {
			lines = append(lines, line)
			line = w
		} else {
			if line == "" {
				line = w
			} else {
				line += " " + w
			}
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n  ")
}

func printSummary(approved, deleted, skipped, remaining int) {
	fmt.Println("══════════════════════════════════════════════════")
	fmt.Printf("  Done — approved: %d  deleted: %d  skipped: %d", approved, deleted, skipped)
	if remaining > 0 {
		fmt.Printf("  (quit early, %d remaining)", remaining)
	}
	fmt.Println()
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "❌  "+format+"\n", args...)
	os.Exit(1)
}
